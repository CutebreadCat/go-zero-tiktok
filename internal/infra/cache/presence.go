package cache

import (
	"context"
	"log"
	"time"
)

func (c *RedisCache) OnlineKey(userID string) string {
	return "presence:" + userID
}

func (c *RedisCache) SetOnline(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	_, err := c.client.Zadd(c.OnlineKey(userID), int64(now), url)
	if err != nil {
		log.Printf("❌ 无法设置在线状态 (userID: %s): %v", userID, err)
		return err
	}

	c.client.Expire(c.OnlineKey(userID), OnlineExpireSeconds)

	log.Printf("✅ 用户 %s 在线，设置在线状态成功！", userID)
	return nil
}

func (c *RedisCache) SetOffline(ctx context.Context, userID string, url string) error {
	_, err := c.client.Zrem(c.OnlineKey(userID), url)
	if err != nil {
		log.Printf("❌ 无法设置离线状态 (userID: %s): %v", userID, err)
		return err
	}

	log.Printf("✅ 用户 %s 离线，设置离线状态成功！", userID)
	return nil
}

func (c *RedisCache) HeartBeat(ctx context.Context, userID string, url string) error {
	now := time.Now().Unix()
	// 1. 更新该设备在 ZSet 中的活跃时间戳
	if _, err := c.client.Zadd(c.OnlineKey(userID), now, url); err != nil {
		log.Printf("❌ 用户 %s 心跳更新 Score 失败：%v", userID, err)
		return err
	}

	// 2. 续命用户在线状态
	if err := c.client.Expire(c.OnlineKey(userID), OnlineExpireSeconds); err != nil {
		log.Printf("❌ 用户 %s 心跳续命 TTL 失败：%v", userID, err)
		return err
	}

	return nil
}

func (c *RedisCache) IsOnline(ctx context.Context, userID string) (bool, error) {
	count, err := c.client.Zcard(c.OnlineKey(userID))
	if err != nil {
		log.Printf("❌ 无法获取在线状态 (userID: %s): %v", userID, err)
		return false, err
	}
	return count > 0, nil
}
