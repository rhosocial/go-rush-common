package component

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type RedisClientPool struct {
	clients *[]*redis.Client
	turnMap []uint8
}

var RedisServerTurn atomic.Uint32

var ErrRedisClientNil = errors.New("redis client nil")
var ErrRedisClientsNotAvailable = errors.New("redis client(s) not available")

// GetCurrentTurn 获得当前活动 redis 客户端顺序。如果没有活动客户端，则报 ErrRedisClientNil 错误。
// 注意！如果某个 Redis 服务器的权重大于1，则意味着该服务器将被询问多次。
func (c *RedisClientPool) GetCurrentTurn() *uint8 {
	if c == nil || c.clients == nil || len(*c.clients) == 0 {
		panic(ErrRedisClientNil)
	}
	now := RedisServerTurn.Load() % uint32(len(c.turnMap))
	next := RedisServerTurn.Add(1) % uint32(len(c.turnMap))
	status := *c.GetRedisServerStatus(context.Background(), c.turnMap[next])
	for {
		if now == next && !status.Valid {
			panic(ErrRedisClientsNotAvailable)
		}
		if now != next && !status.Valid {
			next = RedisServerTurn.Add(1) % uint32(len(c.turnMap))
			status = *c.GetRedisServerStatus(context.Background(), c.turnMap[next])
			continue
		}
		break
	}
	return &c.turnMap[next]
}

// GetCurrentClient 获得当前活动 redis 客户端指针。如果没有活动客户端，则报 ErrRedisClientNil 错误。
func (c *RedisClientPool) GetCurrentClient() *redis.Client {
	if c.GetCurrentTurn() == nil {
		return nil
	}
	return (*c.clients)[*c.GetCurrentTurn()]
}

type EnvRedisServerDialer struct {
	KeepAlive uint8 `yaml:"KeepAlive,omitempty" default:"5" validate:"min=1,max=10"`
	Timeout   uint8 `yaml:"Timeout,omitempty" default:"1" validate:"min=1,max=10"`
}

func (e *EnvRedisServerDialer) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

func (e *EnvRedisServer) GetDialerDefault() *EnvRedisServerDialer {
	d := EnvRedisServerDialer{KeepAlive: 5, Timeout: 1}
	return &d
}

type EnvRedisServerWorker struct {
	Interval uint16 `yaml:"Interval,omitempty" default:"1000" validate:"min=100,max=60000"`
}

func (e *EnvRedisServerWorker) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

func (e *EnvRedisServer) GetWorkerDefault() *EnvRedisServerWorker {
	d := EnvRedisServerWorker{Interval: 1000}
	return &d
}

type EnvRedisServer struct {
	Host     string                `yaml:"Host,omitempty" default:"localhost"`
	Port     uint16                `yaml:"Port,omitempty" default:"6379"`
	Username string                `yaml:"Username,omitempty" default:""`
	Password string                `yaml:"Password,omitempty" default:""`
	DB       int                   `yaml:"DB,omitempty" default:"0" validate:"min=0,max=15"`
	Weight   uint8                 `yaml:"Weight,omitempty" default:"1" validate:"min=1,max=10"`
	Dialer   *EnvRedisServerDialer `yaml:"Dialer,omitempty"`
	Worker   *EnvRedisServerWorker `yaml:"Worker,omitempty"`
}

func (e *EnvRedisServer) Validate() error {
	if e.Dialer != nil {
		if err := e.Dialer.Validate(); err != nil {
			return err
		}
	} else {
		e.Dialer = e.GetDialerDefault()
	}
	if e.Worker != nil {
		if err := e.Worker.Validate(); err != nil {
			return err
		}
	} else {
		e.Worker = e.GetWorkerDefault()
	}
	validate := validator.New()
	return validate.Struct(e)
}

func (e *EnvRedisServer) GetRedisOptions() *redis.Options {
	if e == nil {
		return nil
	}
	options := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", e.Host, e.Port),
		Username: e.Username,
		Password: e.Password,
		DB:       e.DB,
		Dialer: func(ctx context.Context, network, address string) (net.Conn, error) {
			config := e.Dialer
			if config == nil {
				config = e.GetDialerDefault()
			}
			netDialer := &net.Dialer{
				Timeout:   time.Duration(config.Timeout) * time.Second,
				KeepAlive: time.Duration(config.KeepAlive) * time.Minute,
			}
			return netDialer.Dial(network, address)
		},
	}
	return &options
}

func (c *RedisClientPool) InitRedisClientPool(servers *[]EnvRedisServer) {
	redisClients := make([]*redis.Client, len(*servers))
	turnMap := make([]uint8, 0)
	for i, v := range *servers {
		redisClients[i] = redis.NewClient(v.GetRedisOptions())
		for j := 0; j < int(v.Weight); j++ {
			turnMap = append(turnMap, uint8(i))
		}
	}
	c.clients = &redisClients
	c.turnMap = turnMap
}

type RedisServerStatus struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

func (c *RedisClientPool) GetClient(idx *uint8) *redis.Client {
	if c == nil || c.clients == nil {
		panic(ErrRedisClientNil)
	}
	index := uint8(0)
	if idx != nil {
		index = *idx
	}
	if int(index) >= len(*c.clients) || (*c.clients)[index] == nil {
		panic(ErrRedisClientNil)
	}
	return (*c.clients)[index]
}

func (c *RedisClientPool) GetRedisServerStatus(ctx context.Context, idx uint8) *RedisServerStatus {
	client := c.GetClient(&idx)
	if client == nil {
		panic(ErrRedisClientNil)
	}
	poolStats := client.PoolStats()
	status := RedisServerStatus{
		Valid: false,
	}
	if _, err := client.Ping(ctx).Result(); err == nil {
		status.Valid = true
		status.Message = fmt.Sprintf("命中:%d, 未命中:%d, 超时:%d, 总连接:%d, 空闲连接:%d, 失效连接:%d.", poolStats.Hits, poolStats.Misses, poolStats.Timeouts, poolStats.TotalConns, poolStats.IdleConns, poolStats.StaleConns)
	} else {
		status.Valid = false
		status.Message = err.Error()
	}
	return &status
}

func (c *RedisClientPool) GetRedisServersStatus(ctx context.Context) map[uint8]RedisServerStatus {
	result := make(map[uint8]RedisServerStatus)
	for i := 0; c != nil && c.clients != nil && i < len(*c.clients); i++ {
		status := c.GetRedisServerStatus(ctx, uint8(i))
		result[uint8(i)] = *status
	}
	return result
}
