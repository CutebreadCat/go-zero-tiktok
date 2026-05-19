package limiter

import (
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Limiter interface {
	Allow(key string) (bool, error)
}

type PeriodLimiter struct {
	limiter *limit.PeriodLimit
}

func New(rds *redis.Redis, seconds int, maxRequests int, keyPrefix string) Limiter {
	return &PeriodLimiter{
		limiter: limit.NewPeriodLimit(seconds, maxRequests, rds, keyPrefix),
	}
}

func (l *PeriodLimiter) Allow(key string) (bool, error) {
	result, err := l.limiter.Take(key)
	if err != nil {
		return false, err
	}
	return result == limit.Allowed, nil
}
