package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/types"
	"log"
)

func (r *RedisCache) streamkey(roomID string) string {
	return fmt.Sprintf("stream:message:%s", roomID)
}
func (r *RedisCache) Unreadkey(userid, roomid string) string {
	return fmt.Sprintf("unread:%s:%s", userid, roomid)
}

type CacheMessage struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Context   string `json:"context"`
	CreatedAt string `json:"created_at"`
}

func (r *RedisCache) AddMessage(ctx context.Context, message *types.MessageChat) (string, error) {
	cacheMsg := CacheMessage{
		ID:        message.ID,
		RoomID:    message.RoomID,
		SenderID:  message.SenderID,
		Context:   message.Content,
		CreatedAt: message.CreatedAt,
	}
	data, err := json.Marshal(cacheMsg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return "", err
	}
	key := r.streamkey(message.RoomID)
	id, err := r.client.XAdd(key, false, "*", []any{
		"data", data,
	})
	if err != nil {
		log.Printf("Failed to add message to stream: %v", err)
		return "", err
	}
	maxlen := 1000
	_, err = r.client.Do("XTRIM", key, "MAXLEN", maxlen)
	if err != nil {
		log.Printf("裁剪 Stream 长度失败: %v", err)
		return "", err
	} //这里需要后续做一个异步操作，定期裁剪 Stream 长度，避免过长导致性能问题
	return id, nil

}
func (r *RedisCache) IncrUnread(ctx context.Context, userID, roomID string) error {
	key := r.Unreadkey(userID, roomID)
	_, err := r.client.Incr(key)
	if err != nil {
		log.Printf("Failed to increment unread count for user %s in room %s: %v", userID, roomID, err)
		return err
	}
	return nil
}
