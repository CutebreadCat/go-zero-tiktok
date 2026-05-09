package cache

import (
	"context"
	"log"
)

const OnlineExpireSeconds = 24 * 3600 // 24小时

func (c *RedisCache) RoomOnlineKey(roomID string) string {
	return "presence:room:" + roomID
}

// ============ 房间在线用户集合操作 ============

// JoinRoom 将用户添加到房间集合
func (c *RedisCache) JoinRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Sadd(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("❌ 无法将用户 %s 添加到房间 %s: %v", userID, roomID, err)
		return err
	}

	// 设置房间 key 的过期时间
	c.client.Expire(c.RoomOnlineKey(roomID), OnlineExpireSeconds)

	log.Printf("✅ 用户 %s 已加入房间 %s", userID, roomID)
	return nil
}

// LeaveRoom 将用户从房间集合移除
func (c *RedisCache) LeaveRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Srem(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("❌ 无法将用户 %s 从房间 %s 移除: %v", userID, roomID, err)
		return err
	}

	log.Printf("✅ 用户 %s 已离开房间 %s", userID, roomID)
	return nil
}

// GetRoomOnlineUsers 获取房间在线用户列表
func (c *RedisCache) GetRoomOnlineUsers(ctx context.Context, roomID string) ([]string, error) {
	users, err := c.client.Smembers(c.RoomOnlineKey(roomID))
	if err != nil {
		log.Printf("❌ 无法获取房间 %s 在线用户列表: %v", roomID, err)
		return nil, err
	}
	return users, nil
}

// IsUserInRoom 检查用户是否在房间中
func (c *RedisCache) IsUserInRoom(ctx context.Context, roomID string, userID string) (bool, error) {
	exists, err := c.client.Sismember(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("❌ 无法检查用户 %s 是否在房间 %s: %v", userID, roomID, err)
		return false, err
	}
	return exists, nil
}

// GetRoomOnlineCount 获取房间在线用户数量
func (c *RedisCache) GetRoomOnlineCount(ctx context.Context, roomID string) (int64, error) {
	count, err := c.client.Scard(c.RoomOnlineKey(roomID))
	if err != nil {
		log.Printf("❌ 无法获取房间 %s 在线用户数量: %v", roomID, err)
		return 0, err
	}
	return count, nil
}

// RoomHeartBeat 续命房间集合
func (c *RedisCache) RoomHeartBeat(ctx context.Context, roomID string) error {
	if err := c.client.Expire(c.RoomOnlineKey(roomID), OnlineExpireSeconds); err != nil {
		log.Printf("❌ 房间 %s 心跳续命 TTL 失败：%v", roomID, err)
		return err
	}
	return nil
}
