package websocket

import (
	"context"
	"fmt"
	"log"
	"time"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"
)

type MessageHandler struct {
	messages MessageManager
}

func NewMessageHandler(messages MessageManager) *MessageHandler {
	return &MessageHandler{messages: messages}
}

func (h *MessageHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	fmt.Printf("消费消息事件: type=%s, data=%+v\n", e.Type, e.Data)
	event, ok := e.Data.(*MessageEvent)
	if !ok {
		fmt.Printf("事件数据类型无效: %T\n", e.Data)
		return nil
	}

	log.Printf("MQ 消费消息: room=%s, sender=%s", event.RoomID, event.SenderID)

	msg := &types.MessageChat{
		RoomID:    event.RoomID,
		SenderID:  event.SenderID,
		Content:   event.Content,
		CreatedAt: time.Now().Format(dateTimeLayout),
	}
	h.messages.HandleMessageByUserID(ctx, event.SenderID, msg)

	return nil
}

type UnreadHandler struct {
	messages MessageManager
}

func NewUnreadHandler(messages MessageManager) *UnreadHandler {
	return &UnreadHandler{messages: messages}
}

func (h *UnreadHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(*UnreadEvent)
	if !ok {
		fmt.Printf("未读事件数据类型无效: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 消费未读事件: room=%s, user=%s\n", event.RoomID, event.UserID)
	h.messages.HandleGetUnreadByUserID(ctx, event.UserID, event.RoomID)
	return nil
}

type RoomHandler struct {
	rooms RoomManager
}

func NewRoomHandler(rooms RoomManager) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

func (h *RoomHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(*RoomEvent)
	if !ok {
		fmt.Printf("房间事件数据类型无效: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 消费房间事件: action=%s, room=%s, user=%s\n", event.Action, event.RoomID, event.UserID)
	return nil
}

type AIChatHandler struct {
	ai    *AIChat
	rooms RoomManager
}

func NewAIChatHandler(ai *AIChat, rooms RoomManager) *AIChatHandler {
	return &AIChatHandler{ai: ai, rooms: rooms}
}

func (h *AIChatHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(*AIChatEvent)
	if !ok {
		fmt.Printf("AI 聊天事件数据类型无效: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 消费 AI 聊天事件: room=%s, user=%s\n", event.RoomID, event.UserID)

	reply, err := h.ai.ExecuteAI(ctx, event.UserID, event.RoomID)
	if err != nil {
		fmt.Printf("执行 AI 失败 (用户 %s, 房间 %s): %v\n", event.UserID, event.RoomID, err)
		return nil
	}

	if reply.Message.Content != "" {
		h.rooms.BroadcastToRoom(event.RoomID, reply)
	}

	return nil
}
