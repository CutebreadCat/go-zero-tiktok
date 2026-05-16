package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/domain/websocket"
	"go_zero-tiktok/internal/types"
	"log"

	strconv "strconv"
)

func (r *RedisCache) streamkey(roomID string) string {
	return fmt.Sprintf("stream:message:%s", roomID)
}
func (r *RedisCache) Unreadkey(userid, roomid string) string {
	return fmt.Sprintf("unread:%s:%s", userid, roomid)
}
func (r *RedisCache) AiChatKey(roomID, userid string) string {
	return fmt.Sprintf("aichat:%s:%s", userid, roomID)
}

func (r *RedisCache) AddMessage(ctx context.Context, message *types.MessageChat) (string, error) {
	cacheMsg := websocket.CacheMessage{
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
		log.Printf("Failed to get unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}

	if count == "" {
		return 0, nil
	}

	countInt, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		log.Printf("Failed to parse unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}
	return countInt, nil

}

func (r *RedisCache) GetMessages(ctx context.Context, userID, roomID string, count int64) ([]websocket.CacheMessage, error) {
	messagekey := r.streamkey(roomID)

	// 使用 Do 方法执行 XREVRANGE 命令
	result, err := r.client.Do("XREVRANGE", messagekey, "+", "-", "COUNT", count)
	if err != nil {
		log.Printf("Failed to get messages from stream: %v", err)
		return nil, err
	}

	// 解析返回结果
	// XREVRANGE 返回的是 []interface{}，每个元素是 [id, [field, value, ...]]
	resultSlice, ok := result.([]interface{})
	if !ok {
		log.Printf("Failed to parse XREVRANGE result type: %T", result)
		return nil, nil
	}

	var messages []websocket.CacheMessage
	for _, item := range resultSlice {
		msgSlice, ok := item.([]interface{})
		if !ok || len(msgSlice) < 2 {
			continue
		}

		// msgSlice[0] 是消息 ID，msgSlice[1] 是字段值对
		fields, ok := msgSlice[1].([]interface{})
		if !ok {
			continue
		}

		// 遍历字段值对，找到 "data" 字段
		for i := 0; i < len(fields)-1; i += 2 {
			fieldName, ok := fields[i].(string)
			if !ok || fieldName != "data" {
				continue
			}
			data, ok := fields[i+1].(string)
			if !ok {
				continue
			}
			var cacheMsg websocket.CacheMessage
			if err := json.Unmarshal([]byte(data), &cacheMsg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			messages = append(messages, cacheMsg)
		}
	}

	// 反转消息顺序（XREVRANGE 返回的是从新到旧，我们需要从旧到新）
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

func (r *RedisCache) IncrAIMessage(ctx context.Context, userID, roomID string) error {
	key := r.AiChatKey(roomID, userID)
	_, err := r.client.Incr(key)
	if err != nil {
		log.Printf("Failed to increment unread count for user %s in room %s: %v", userID, roomID, err)
		return err
	}
	return nil
}
func (r *RedisCache) GetAIMessageCount(ctx context.Context, userID, roomID string) (int64, error) {
	key := r.AiChatKey(roomID, userID)

	count, err := r.client.Get(key)
	if err != nil {
		log.Printf("Failed to get unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}

	if count == "" {
		return 0, nil
	}

	countInt, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		log.Printf("Failed to parse unread count for user %s in room %s: %v", userID, roomID, err)
		return 0, err
	}
	return countInt, nil

}

func (r *RedisCache) ClearAIMessage(ctx context.Context, userID, roomID string) error {
	key := r.AiChatKey(roomID, userID)
	_, err := r.client.Del(key)
	if err != nil {
		log.Printf("Failed to clear AI message count for user %s in room %s: %v", userID, roomID, err)
		return err
	}
	return nil
}
