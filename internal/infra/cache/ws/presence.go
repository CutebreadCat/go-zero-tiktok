package ws

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"
	"log"
	"time"
)

func (c *RedisCache) OnlineKey(userID string) string {
	return presenceKeyPrefix + userID
}

func (c *RedisCache) SetOnline(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	_, err := c.client.Zadd(c.OnlineKey(userID), int64(now), url)
	if err != nil {
		log.Printf("无法设置在线状态 (userID: %s): %v", userID, err)
		return xerr.Wrap(err, "RedisCache.SetOnline")
	}

	c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds)

	log.Printf("用户 %s 在线，设置在线状态成功", userID)
	return nil
}

func (c *RedisCache) SetOffline(ctx context.Context, userID string, url string) error {
	_, err := c.client.Zrem(c.OnlineKey(userID), url)
	if err != nil {
		log.Printf("无法设置离线状态 (userID: %s): %v", userID, err)
		return xerr.Wrap(err, "RedisCache.SetOffline")
	}

	log.Printf("用户 %s 离线，设置离线状态成功", userID)
	return nil
}

func (c *RedisCache) HeartBeat(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	if _, err := c.client.Zadd(c.OnlineKey(userID), now, url); err != nil {
		log.Printf("用户 %s 心跳更新 Score 失败: %v", userID, err)
		return xerr.Wrap(err, "RedisCache.HeartBeat.Zadd")
	}

	if err := c.client.Expire(c.OnlineKey(userID), onlineExpireSeconds); err != nil {
		log.Printf("用户 %s 心跳续命 TTL 失败: %v", userID, err)
		return xerr.Wrap(err, "RedisCache.HeartBeat.Expire")
	}

	return nil
}

func (c *RedisCache) IsOnline(ctx context.Context, userID string) (bool, error) {
	count, err := c.client.Zcard(c.OnlineKey(userID))
	if err != nil {
		log.Printf("无法获取在线状态 (userID: %s): %v", userID, err)
		return false, xerr.Wrap(err, "RedisCache.IsOnline")
	}
	return count > 0, nil
}
