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

type messageManager struct {
	cache    MessageCache
	repo     MessageRepository
	roomRepo RoomRepository
	rooms    RoomManager
	writer   MessageWriter
	ai       *AIChat
}

func NewMessageManager(cache MessageCache, repo MessageRepository, roomRepo RoomRepository, rooms RoomManager, writer MessageWriter, ai *AIChat) MessageManager {
	return &messageManager{
		cache:    cache,
		repo:     repo,
		roomRepo: roomRepo,
		rooms:    rooms,
		writer:   writer,
		ai:       ai,
	}
}

func (mm *messageManager) HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat) {
	if !mm.rooms.IsMember(client, msg.RoomID) {
		log.Printf("用户 %s 不是房间 %s 的成员，忽略消息", client.UserID, msg.RoomID)
		return
	}

	broadcastMsg := Message{
		Message: *msg,
		Typek:   "message",
	}
	mm.rooms.BroadcastToRoom(msg.RoomID, broadcastMsg)

	event := NewMessageEvent(MessageTopic, msg.RoomID, client.UserID, msg.Content)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		log.Printf("发送消息事件到 Kafka 失败 (用户 %s, 房间 %s): %v", client.UserID, msg.RoomID, err)
	}

	go func(userID, roomID, content string) {
		reached, err := mm.ai.CheckAndEnqueue(ctx, userID, roomID, msg)
		if err != nil {
			log.Printf("检查 AI 入队失败 (用户 %s, 房间 %s): %v", userID, roomID, err)
			return
		}
		if reached {
			aiEvent := NewAIChatEvent(MessageTopic, roomID, userID, content)
			if err := mm.writer.SendMessage(ctx, aiEvent); err != nil {
				log.Printf("发送 AI 聊天事件到 Kafka 失败 (用户 %s, 房间 %s): %v", userID, roomID, err)
			}
		}
	}(client.UserID, msg.RoomID, msg.Content)
}

func (mm *messageManager) HandleMessageByUserID(ctx context.Context, userID string, msg *types.MessageChat) {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	if _, err := mm.cache.AddMessage(ctx, msg); err != nil {
		log.Printf("添加消息到缓存失败 (房间 %s): %v", msg.RoomID, err)
	}

	if err := mm.repo.StoreChatMessage(ctx, msg); err != nil {
		log.Printf("存储消息到数据库失败 (用户 %s, 房间 %s): %v", userID, msg.RoomID, err)
	}

	mm.UpdateUnreadCount(ctx, userID, msg.RoomID)

}
