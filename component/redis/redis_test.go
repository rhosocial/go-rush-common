package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisClientPool_GetRedisServersStatus(t *testing.T) {
	t.Run("nil RedisClientPool", func(t *testing.T) {
		pool := RedisClientPool{}
		result := pool.GetRedisServersStatus(context.Background())
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})
}
