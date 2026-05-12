package websocket

import (
	"context"

	"log"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"
)

// MessageHandler 处理消息事件（MQ 消费端）
type MessageHandler struct {
	messages MessageManager
}

func NewMessageHandler(messages MessageManager) *MessageHandler {
	return &MessageHandler{messages: messages}
}

func (h *MessageHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(MessageEvent)
	if !ok {
		log.Printf("Invalid event data type: %T", e.Data)
		return nil
	}

	log.Printf("MQ 消费消息事件: room=%s, sender=%s", event.RoomID, event.SenderID)

	msg := &types.MessageChat{
		RoomID:   event.RoomID,
		SenderID: event.SenderID,
		Content:  event.Content,
	}
	h.messages.HandleMessageByUserID(ctx, event.SenderID, msg)
	return nil
}

// UnreadHandler 处理未读消息事件（MQ 消费端）
type UnreadHandler struct {
	messages MessageManager
}

func NewUnreadHandler(messages MessageManager) *UnreadHandler {
	return &UnreadHandler{messages: messages}
}

func (h *UnreadHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(UnreadEvent)
	if !ok {
		log.Printf("Invalid event data type: %T", e.Data)
		return nil
	}

	log.Printf("MQ 消费未读消息事件: room=%s, user=%s", event.RoomID, event.UserID)
	h.messages.HandleGetUnreadByUserID(ctx, event.UserID, event.RoomID)
	return nil
}

// RoomHandler 处理房间事件（MQ 消费端）
type RoomHandler struct {
	rooms RoomManager
}

func NewRoomHandler(rooms RoomManager) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

func (h *RoomHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	event, ok := e.Data.(RoomEvent)
	if !ok {
		log.Printf("Invalid event data type: %T", e.Data)
		return nil
	}

	log.Printf("MQ 消费房间事件: action=%s, room=%s, user=%s", event.Action, event.RoomID, event.UserID)
	return nil
}
