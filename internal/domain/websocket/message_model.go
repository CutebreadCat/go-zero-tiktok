package websocket

import (
	"context"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"
)

// ==================== 接口定义 ====================

// MessageCache 消息缓存接口
type MessageCache interface {
	AddMessage(ctx context.Context, message *types.MessageChat) (string, error)
	IncrUnread(ctx context.Context, userID, roomID string) error
	GetUnreadMessages(ctx context.Context, userID, roomID string, count int64) ([]CacheMessage, error)
	GetUnreadCount(ctx context.Context, userID, roomID string) (int64, error)
	ClearUnread(ctx context.Context, userID, roomID string) error
}

// MessageRepository 消息持久化接口
type MessageRepository interface {
	StoreChatMessage(ctx context.Context, message *types.MessageChat) error
}

// MessageWriter 消息写入器接口（依赖注入 Kafka Producer）
type MessageWriter interface {
	SendMessage(ctx context.Context, event *mqcontract.Event) error
}

// MessageManager 消息管理接口
type MessageManager interface {
	HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat)
	HandleGetUnread(ctx context.Context, client *Client, roomID string)
	HandleMessageByUserID(ctx context.Context, userID string, msg *types.MessageChat)
	HandleGetUnreadByUserID(ctx context.Context, userID, roomID string)
}

// ==================== 模型定义 ====================

// UnreadResponse 未读消息响应
type UnreadResponse struct {
	Typek    string         `json:"typek"`
	RoomID   string         `json:"room_id"`
	Count    int64          `json:"count"`
	Messages []CacheMessage `json:"messages"`
}

type CacheMessage struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Context   string `json:"context"`
	CreatedAt string `json:"created_at"`
}
