package ws

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"
	"time"
)

func (c *RedisCache) OnlineKey(userID string) string {
	return presenceKeyPrefix + userID
}

func (c *RedisCache) SetOnline(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	_, err := c.client.Zadd(c.OnlineKey(userID), int64(now), url)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.SetOnline")
	}

	c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds)

	fmt.Printf("用户 %s 在线，设置在线状态成功\n", userID)
	return nil
}

func (c *RedisCache) SetOffline(ctx context.Context, userID string, url string) error {
	_, err := c.client.Zrem(c.OnlineKey(userID), url)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.SetOffline")
	}

	fmt.Printf("用户 %s 离线，设置离线状态成功\n", userID)
	return nil
}

func (c *RedisCache) HeartBeat(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	if _, err := c.client.Zadd(c.OnlineKey(userID), now, url); err != nil {
		return xerr.Wrap(err, "RedisCache.HeartBeat.Zadd")
	}

	if err := c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds); err != nil {
		return xerr.Wrap(err, "RedisCache.HeartBeat.Expire")
	}

	return nil
}

func (c *RedisCache) IsOnline(ctx context.Context, userID string) (bool, error) {
	count, err := c.client.Zcard(c.OnlineKey(userID))
	if err != nil {
		return false, xerr.Wrap(err, "RedisCache.IsOnline")
	}
	return count > 0, nil
}
