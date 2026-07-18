package ws

import (
	"context"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/pkg/xerr"
)

func (c *RedisCache) RoomOnlineKey(roomID string) string {
	return roomPresenceKeyPrefix + roomID
}

func (c *RedisCache) JoinRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Sadd(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		appLogger.Infof("鏃犳硶灏嗙敤鎴?%s 娣诲姞鍒版埧闂?%s: %v", userID, roomID, err)
		return xerr.Wrap(err, "RedisCache.JoinRoom")
	}

	_ = c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds)

	appLogger.Infof("鐢ㄦ埛 %s 宸插姞鍏ユ埧闂?%s", userID, roomID)
	return nil
}

func (c *RedisCache) LeaveRoom(ctx context.Context, roomID string, userID string) error {
	_, err := c.client.Srem(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		appLogger.Infof("鏃犳硶灏嗙敤鎴?%s 浠庢埧闂?%s 绉婚櫎: %v", userID, roomID, err)
		return xerr.Wrap(err, "RedisCache.LeaveRoom")
	}

	appLogger.Infof("鐢ㄦ埛 %s 宸茬寮€鎴块棿 %s", userID, roomID)
	return nil
}

func (c *RedisCache) GetRoomOnlineUsers(ctx context.Context, roomID string) ([]string, error) {
	users, err := c.client.Smembers(c.RoomOnlineKey(roomID))
	if err != nil {
		appLogger.Infof("鏃犳硶鑾峰彇鎴块棿 %s 鍦ㄧ嚎鐢ㄦ埛鍒楄〃: %v", roomID, err)
		return nil, xerr.Wrap(err, "RedisCache.GetRoomOnlineUsers")
	}
	return users, nil
}

func (c *RedisCache) IsUserInRoom(ctx context.Context, roomID string, userID string) (bool, error) {
	exists, err := c.client.Sismember(c.RoomOnlineKey(roomID), userID)
	if err != nil {
		appLogger.Infof("鏃犳硶妫€鏌ョ敤鎴?%s 鏄惁鍦ㄦ埧闂?%s: %v", userID, roomID, err)
		return false, xerr.Wrap(err, "RedisCache.IsUserInRoom")
	}
	return exists, nil
}

func (c *RedisCache) GetRoomOnlineCount(ctx context.Context, roomID string) (int64, error) {
	count, err := c.client.Scard(c.RoomOnlineKey(roomID))
	if err != nil {
		appLogger.Infof("鏃犳硶鑾峰彇鎴块棿 %s 鍦ㄧ嚎鐢ㄦ埛鏁伴噺: %v", roomID, err)
		return 0, xerr.Wrap(err, "RedisCache.GetRoomOnlineCount")
	}
	return count, nil
}

func (c *RedisCache) RoomHeartBeat(ctx context.Context, roomID string) error {
	if err := c.client.Expire(c.RoomOnlineKey(roomID), onlineExpireSeconds); err != nil {
		appLogger.Infof("鎴块棿 %s 蹇冭烦缁懡 TTL 澶辫触: %v", roomID, err)
		return xerr.Wrap(err, "RedisCache.RoomHeartBeat")
	}
	return nil
}
