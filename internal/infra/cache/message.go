package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/types"
	"log"

	strconv "strconv"

	"github.com/redis/go-redis/v9"
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
func (r *RedisCache) GetUnreadCount(ctx context.Context, userID, roomID string) (int64, error) {
	key := r.Unreadkey(userID, roomID)

	count, err := r.client.Get(key)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		log.Printf("Failed to get unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}
	countInt, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		log.Printf("Failed to parse unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}
	return countInt, nil

}

func (r *RedisCache) GetUnreadMessages(ctx context.Context, userID, roomID string, count int64) ([]CacheMessage, error) {
	messagekey := r.streamkey(roomID)

	var msgs []redis.XMessage
	_, err := r.client.Do(&msgs, "XREVRANGE", messagekey, "+", "-", "COUNT", count)
	if err != nil {
		log.Printf("Failed to get messages from stream: %v", err)
		return nil, err
	}
	var messages []CacheMessage
	for _, msg := range msgs {
		var cacheMsg CacheMessage
		data, ok := msg.Values["data"].(string)
		if !ok {
			log.Printf("Failed to assert message data type")
			continue
		}
		if err := json.Unmarshal([]byte(data), &cacheMsg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		messages = append(messages, cacheMsg)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}
func (r *RedisCache) ClearUnread(ctx context.Context, userID, roomID string) error {
	key := r.Unreadkey(userID, roomID)
	_, err := r.client.Del(key)
	if err != nil {
		log.Printf("Failed to clear unread count for user %s in room %s: %v", userID, roomID, err)
		return err
	}
	return nil
}
