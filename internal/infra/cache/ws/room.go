package ws

import (
	"context"
	"go_zero-tiktok/internal/shared/xerr"
	"log"
)

func (c *RedisCache) RoomOnlineKey(roomID string) string {
	return roomPresenceKeyPrefix + roomID
}

func (c *RedisCache) JoinRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Sadd(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("无法将用户 %s 添加到房间 %s: %v", userID, roomID, err)
		return xerr.Wrap(err, "RedisCache.JoinRoom")
	}

	_ = c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds)

	log.Printf("用户 %s 已加入房间 %s", userID, roomID)
	return nil
}

func (c *RedisCache) LeaveRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Srem(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("无法将用户 %s 从房间 %s 移除: %v", userID, roomID, err)
		return xerr.Wrap(err, "RedisCache.LeaveRoom")
	}

	log.Printf("用户 %s 已离开房间 %s", userID, roomID)
	return nil
}

func (c *RedisCache) GetRoomOnlineUsers(ctx context.Context, roomID string) ([]string, error) {
	users, err := c.client.Smembers(c.RoomOnlineKey(roomID))
	if err != nil {
		log.Printf("无法获取房间 %s 在线用户列表: %v", roomID, err)
		return nil, xerr.Wrap(err, "RedisCache.GetRoomOnlineUsers")
	}
	return users, nil
}

func (c *RedisCache) IsUserInRoom(ctx context.Context, roomID string, userID string) (bool, error) {
	exists, err := c.client.Sismember(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		log.Printf("无法检查用户 %s 是否在房间 %s: %v", userID, roomID, err)
		return false, xerr.Wrap(err, "RedisCache.IsUserInRoom")
	}
	return exists, nil
}

func (c *RedisCache) GetRoomOnlineCount(ctx context.Context, roomID string) (int64, error) {
	count, err := c.client.Scard(c.RoomOnlineKey(roomID))
	if err != nil {
		log.Printf("无法获取房间 %s 在线用户数量: %v", roomID, err)
		return 0, xerr.Wrap(err, "RedisCache.GetRoomOnlineCount")
	}
	return count, nil
}

func (c *RedisCache) RoomHeartBeat(ctx context.Context, roomID string) error {
	if err := c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds); err != nil {
		log.Printf("房间 %s 心跳续命 TTL 失败: %v", roomID, err)
		return xerr.Wrap(err, "RedisCache.RoomHeartBeat")
	}
	return nil
}
