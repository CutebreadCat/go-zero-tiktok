package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/domain/websocket"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"

	strconv "strconv"
)

func (r *RedisCache) streamkey(roomID string) string {
	return messageStreamPrefix + roomID
}
func (r *RedisCache) aiStreamKey(roomID, userID string) string {
	return fmt.Sprintf("%s%s:%s", aiStreamPrefix, roomID, userID)
}
func (r *RedisCache) Unreadkey(userid, roomid string) string {
	return fmt.Sprintf("%s%s:%s", unreadKeyPrefix, userid, roomid)
}
func (r *RedisCache) AiChatKey(userid, roomID string) string {
	return fmt.Sprintf("%s%s:%s", aiChatKeyPrefix, userid, roomID)
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
		return "", xerr.Wrap(err, "RedisCache.AddMessage.Marshal")
	}
	key := r.streamkey(message.RoomID)
	id, err := r.client.XAdd(key, false, "*", []any{
		"data", data,
	})
	if err != nil {
		return "", xerr.Wrap(err, "RedisCache.AddMessage.XAdd")
	}
	_, err = r.client.Do("XTRIM", key, "MAXLEN", messageStreamMaxLen)
	if err != nil {
		return "", xerr.Wrap(err, "RedisCache.AddMessage.XTRIM")
	}
	return id, nil

}
func (r *RedisCache) IncrUnread(ctx context.Context, userID, roomID string) error {
	key := r.Unreadkey(userID, roomID)
	_, err := r.client.Incr(key)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.IncrUnread")
	}
	return nil
}
func (r *RedisCache) GetUnreadCount(ctx context.Context, userID, roomID string) (int64, error) {
	key := r.Unreadkey(userID, roomID)

	count, err := r.client.Get(key)
	if err != nil {
		return 0, xerr.Wrap(err, "RedisCache.GetUnreadCount.Get")
	}

	if count == "" {
		return 0, nil
	}

	countInt, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		return 0, xerr.Wrap(err, "RedisCache.GetUnreadCount.ParseInt")
	}
	return countInt, nil

}

func (r *RedisCache) GetMessages(ctx context.Context, userID, roomID string, count int64) ([]websocket.CacheMessage, error) {
	messagekey := r.streamkey(roomID)

	result, err := r.client.Do("XREVRANGE", messagekey, "+", "-", "COUNT", count)
	if err != nil {
		return nil, xerr.Wrap(err, "RedisCache.GetMessages.XREVRANGE")
	}

	resultSlice, ok := result.([]interface{})
	if !ok {
		return nil, nil
	}

	var messages []websocket.CacheMessage
	for _, item := range resultSlice {
		msgSlice, ok := item.([]interface{})
		if !ok || len(msgSlice) < 2 {
			continue
		}

		fields, ok := msgSlice[1].([]interface{})
		if !ok {
			continue
		}

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
				continue
			}
			messages = append(messages, cacheMsg)
		}
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
		return xerr.Wrap(err, "RedisCache.ClearUnread")
	}
	return nil
}

func (r *RedisCache) IncrAIMessage(ctx context.Context, userID, roomID string) error {
	key := r.AiChatKey(userID, roomID)
	_, err := r.client.Incr(key)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.IncrAIMessage")
	}
	return nil
}
func (r *RedisCache) GetAIMessageCount(ctx context.Context, userID, roomID string) (int64, error) {
	key := r.AiChatKey(userID, roomID)

	count, err := r.client.Get(key)
	if err != nil {
		return 0, xerr.Wrap(err, "RedisCache.GetAIMessageCount.Get")
	}

	if count == "" {
		return 0, nil
	}

	countInt, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		return 0, xerr.Wrap(err, "RedisCache.GetAIMessageCount.ParseInt")
	}
	return countInt, nil

}

func (r *RedisCache) ClearAIMessage(ctx context.Context, userID, roomID string) error {
	key := r.AiChatKey(userID, roomID)
	_, err := r.client.Del(key)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.ClearAIMessage")
	}
	return nil
}

func (r *RedisCache) AddAIMessage(ctx context.Context, userID, roomID string, message *types.MessageChat) (string, error) {
	cacheMsg := websocket.CacheMessage{
		ID:        message.ID,
		RoomID:    message.RoomID,
		SenderID:  message.SenderID,
		Context:   message.Content,
		CreatedAt: message.CreatedAt,
	}
	data, err := json.Marshal(cacheMsg)
	if err != nil {
		return "", xerr.Wrap(err, "RedisCache.AddAIMessage.Marshal")
	}
	key := r.aiStreamKey(roomID, userID)
	id, err := r.client.XAdd(key, false, "*", []any{
		"data", data,
	})
	if err != nil {
		return "", xerr.Wrap(err, "RedisCache.AddAIMessage.XAdd")
	}
	return id, nil
}

func (r *RedisCache) GetAIMessages(ctx context.Context, userID, roomID string, count int64) ([]websocket.CacheMessage, error) {
	key := r.aiStreamKey(roomID, userID)

	result, err := r.client.Do("XREVRANGE", key, "+", "-", "COUNT", count)
	if err != nil {
		return nil, xerr.Wrap(err, "RedisCache.GetAIMessages.XREVRANGE")
	}

	resultSlice, ok := result.([]interface{})
	if !ok {
		return nil, nil
	}

	var messages []websocket.CacheMessage
	for _, item := range resultSlice {
		msgSlice, ok := item.([]interface{})
		if !ok || len(msgSlice) < 2 {
			continue
		}

		fields, ok := msgSlice[1].([]interface{})
		if !ok {
			continue
		}

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
				continue
			}
			messages = append(messages, cacheMsg)
		}
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *RedisCache) ClearAIStream(ctx context.Context, userID, roomID string) error {
	key := r.aiStreamKey(roomID, userID)
	_, err := r.client.Del(key)
	if err != nil {
		return xerr.Wrap(err, "RedisCache.ClearAIStream")
	}
	return nil
}
