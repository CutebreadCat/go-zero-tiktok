package websocket

import (
	"context"
	"log"

	"go_zero-tiktok/internal/types"

	"github.com/google/uuid"
)

const (
	MessageTopic = "chat-messages"
)

// ==================== MessageManager 实现 ====================

type messageManager struct {
	cache    MessageCache
	repo     MessageRepository
	roomRepo RoomRepository
	rooms    RoomManager
	writer   MessageWriter
}

func NewMessageManager(cache MessageCache, repo MessageRepository, roomRepo RoomRepository, rooms RoomManager, writer MessageWriter) MessageManager {
	return &messageManager{
		cache:    cache,
		repo:     repo,
		roomRepo: roomRepo,
		rooms:    rooms,
		writer:   writer,
	}
}

func (mm *messageManager) HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat) {
	if !mm.rooms.IsMember(client, msg.RoomID) {
		log.Printf("User %s is not a member of room %s, ignoring message", client.UserID, msg.RoomID)
		return
	}

	// 1. 立即广播（实时）
	broadcastMsg := Message{
		Message: *msg,
		Typek:   "message",
	}
	mm.rooms.BroadcastToRoom(msg.RoomID, broadcastMsg)

	// 2. 异步：发送到 MQ 做持久化 + 更新未读
	event := NewMessageEvent(MessageTopic, msg.RoomID, client.UserID, msg.Content)
	log.Printf("Sending message event to Kafka: %+v", event)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		log.Printf("Failed to send message event for user %s in room %s: %v", client.UserID, msg.RoomID, err)
	} else {
		log.Printf("Message event sent to Kafka for user %s in room %s", client.UserID, msg.RoomID)
	}
}

func (mm *messageManager) HandleMessageByUserID(ctx context.Context, userID string, msg *types.MessageChat) {
	log.Printf("HandleMessageByUserID called, userID: %s, msg: %+v", userID, msg)

	// 生成消息 ID（如果为空）
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// 1. 添加到缓存
	if _, err := mm.cache.AddMessage(ctx, msg); err != nil {
		log.Printf("Failed to add message to cache for room %s: %v", msg.RoomID, err)
	} else {
		log.Printf("Message added to cache for room %s", msg.RoomID)
	}

	// 2. 持久化到数据库
	if err := mm.repo.StoreChatMessage(ctx, msg); err != nil {
		log.Printf("Failed to store message from user %s in room %s: %v", userID, msg.RoomID, err)
	} else {
		log.Printf("Message stored to database for room %s", msg.RoomID)
	}

	// 3. 更新未读计数
	mm.UpdateUnreadCount(ctx, userID, msg.RoomID)
}
