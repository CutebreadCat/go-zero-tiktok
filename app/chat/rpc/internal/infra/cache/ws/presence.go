package ws

import (
	"context"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/pkg/xerr"
	"time"
)

func (c *RedisCache) OnlineKey(userID string) string {
	return presenceKeyPrefix + userID
}

func (c *RedisCache) SetOnline(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	_, err := c.client.Zadd(c.OnlineKey(userID), int64(now), url)
	if err != nil {
		appLogger.Errorf("set online status failed: user=%s err=%v", userID, err)
		return xerr.Wrap(err, "RedisCache.SetOnline")
	}

	_ = c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds)

	appLogger.Infof("user %s online", userID)
	return nil
}

func (c *RedisCache) SetOffline(ctx context.Context, userID string, url string) error {
	_, err := c.client.Zrem(c.OnlineKey(userID), url)
	if err != nil {
		appLogger.Errorf("set offline status failed: user=%s err=%v", userID, err)
		return xerr.Wrap(err, "RedisCache.SetOffline")
	}

	appLogger.Infof("user %s offline", userID)
	return nil
}

func (c *RedisCache) HeartBeat(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	if _, err := c.client.Zadd(c.OnlineKey(userID), now, url); err != nil {
		appLogger.Infof("鐢ㄦ埛 %s 蹇冭烦鏇存柊 Score 澶辫触: %v", userID, err)
		return xerr.Wrap(err, "RedisCache.HeartBeat.Zadd")
	}

	if err := c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds); err != nil {
		appLogger.Infof("鐢ㄦ埛 %s 蹇冭烦缁懡 TTL 澶辫触: %v", userID, err)
		return xerr.Wrap(err, "RedisCache.HeartBeat.Expire")
	}

	return nil
}

func (c *RedisCache) IsOnline(ctx context.Context, userID string) (bool, error) {
	count, err := c.client.Zcard(c.OnlineKey(userID))
	if err != nil {
		appLogger.Infof("鏃犳硶鑾峰彇鍦ㄧ嚎鐘舵€?(userID: %s): %v", userID, err)
		return false, xerr.Wrap(err, "RedisCache.IsOnline")
	}
	return count > 0, nil
}
