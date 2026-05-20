package cache

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RedisCache struct {
	client *redis.Redis
}

func NewRedisCache(client *redis.Redis) *RedisCache {
	return &RedisCache{
		client: client,
	}
}
