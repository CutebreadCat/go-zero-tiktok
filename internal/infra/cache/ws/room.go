package ws

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"
)

func (c *RedisCache) RoomOnlineKey(roomID string) string {
	return roomPresenceKeyPrefix + roomID
}

func (c *RedisCache) JoinRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Sadd(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.JoinRoom")
	}

	c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds)

	fmt.Printf("用户 %s 已加入房间 %s\n", userID, roomID)
	return nil
}

func (c *RedisCache) LeaveRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Srem(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.LeaveRoom")
	}

	fmt.Printf("用户 %s 已离开房间 %s\n", userID, roomID)
	return nil
}

func (c *RedisCache) GetRoomOnlineUsers(ctx context.Context, roomID string) ([]string, error) {
	users, err := c.client.Smembers(c.RoomOnlineKey(roomID))
	if err != nil {
		return nil, xerr.Wrap(err, "RedisCache.GetRoomOnlineUsers")
	}
	return users, nil
}

func (c *RedisCache) IsUserInRoom(ctx context.Context, roomID string, userID string) (bool, error) {
	exists, err := c.client.Sismember(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		return false, xerr.Wrap(err, "RedisCache.IsUserInRoom")
	}
	return exists, nil
}

func (c *RedisCache) GetRoomOnlineCount(ctx context.Context, roomID string) (int64, error) {
	count, err := c.client.Scard(c.RoomOnlineKey(roomID))
	if err != nil {
		return 0, xerr.Wrap(err, "RedisCache.GetRoomOnlineCount")
	}
	return count, nil
}

func (c *RedisCache) RoomHeartBeat(ctx context.Context, roomID string) error {
	if err := c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds); err != nil {
		return xerr.Wrap(err, "RedisCache.RoomHeartBeat")
	}
	return nil
}
