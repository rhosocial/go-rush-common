package component

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net"
	"sync/atomic"
	"time"
)

type RedisClientPool struct {
	clients *[]*redis.Client
	turnMap []uint8
}

var RedisServerTurn atomic.Uint32

func (c *RedisClientPool) GetCurrentTurn() *uint8 {
	if c.clients == nil || len(*c.clients) == 0 {
		return nil
	}
	return &c.turnMap[RedisServerTurn.Add(1)%uint32(len(c.turnMap))]
}

func (c *RedisClientPool) GetCurrentClient() *redis.Client {
	if c.GetCurrentTurn() == nil {
		return nil
	}
	return (*c.clients)[*c.GetCurrentTurn()]
}

type EnvRedisServerDialer struct {
	KeepAlive uint8 `yaml:"KeepAlive,omitempty" default:"5"`
	Timeout   uint8 `yaml:"Timeout,omitempty" default:"1"`
}

func (e *EnvRedisServer) GetDialerDefault() *EnvRedisServerDialer {
	d := EnvRedisServerDialer{KeepAlive: 5, Timeout: 1}
	return &d
}

type EnvRedisServerWorker struct {
	Interval uint8 `yaml:"Interval,omitempty" default:"1"`
}

func (e *EnvRedisServer) GetWorkerDefault() *EnvRedisServerWorker {
	d := EnvRedisServerWorker{Interval: 1}
	return &d
}

type EnvRedisServer struct {
	Host     string                `yaml:"Host,omitempty" default:"localhost"`
	Port     uint16                `yaml:"Port,omitempty" default:"6379"`
	Username string                `yaml:"Username,omitempty" default:""`
	Password string                `yaml:"Password,omitempty" default:""`
	DB       int                   `yaml:"DB,omitempty" default:"0"`
	Weight   uint8                 `yaml:"Weight,omitempty" default:"1"`
	Dialer   *EnvRedisServerDialer `yaml:"Dialer,omitempty"`
	Worker   *EnvRedisServerWorker `yaml:"Worker,omitempty"`
}

func (e *EnvRedisServer) GetRedisOptions() *redis.Options {
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

func (c *RedisClientPool) GetRedisServerStatus(ctx context.Context) map[uint8]RedisServerStatus {
	result := make(map[uint8]RedisServerStatus)
	for i, c := range *c.clients {
		status := RedisServerStatus{
			Valid: false,
		}
		poolStats := c.PoolStats()
		if _, err := c.Ping(ctx).Result(); err != nil {
			status.Valid = false
			status.Message = err.Error()
		} else {
			status.Valid = true
			status.Message = fmt.Sprintf("命中:%d, 未命中:%d, 超时:%d, 总连接:%d, 空闲连接:%d, 失效连接:%d.", poolStats.Hits, poolStats.Misses, poolStats.Timeouts, poolStats.TotalConns, poolStats.IdleConns, poolStats.StaleConns)
		}
		result[uint8(i)] = status
	}
	return result
}
