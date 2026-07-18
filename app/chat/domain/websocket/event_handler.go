package websocket

import (
	"context"
	"fmt"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"time"

	"go_zero-tiktok/pkg/contract"
	mqcontract "go_zero-tiktok/pkg/mq"
)

type MessageHandler struct {
	messages MessageManager
}

func NewMessageHandler(messages MessageManager) *MessageHandler {
	return &MessageHandler{messages: messages}
}

func (h *MessageHandler) Consume(ctx context.Context, e *mqcontract.Event) error {
	fmt.Printf("娑堣垂娑堟伅浜嬩欢: type=%s, data=%+v\n", e.Type, e.Data)
	event, ok := e.Data.(*MessageEvent)
	if !ok {
		fmt.Printf("浜嬩欢鏁版嵁绫诲瀷鏃犳晥: %T\n", e.Data)
		return nil
	}

	appLogger.Infof("MQ 娑堣垂娑堟伅: room=%s, sender=%s", event.RoomID, event.SenderID)

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
		fmt.Printf("鏈浜嬩欢鏁版嵁绫诲瀷鏃犳晥: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 娑堣垂鏈浜嬩欢: room=%s, user=%s\n", event.RoomID, event.UserID)
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
		fmt.Printf("鎴块棿浜嬩欢鏁版嵁绫诲瀷鏃犳晥: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 娑堣垂鎴块棿浜嬩欢: action=%s, room=%s, user=%s\n", event.Action, event.RoomID, event.UserID)
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
		fmt.Printf("AI 鑱婂ぉ浜嬩欢鏁版嵁绫诲瀷鏃犳晥: %T\n", e.Data)
		return nil
	}

	fmt.Printf("MQ 娑堣垂 AI 鑱婂ぉ浜嬩欢: room=%s, user=%s\n", event.RoomID, event.UserID)

	reply, err := h.ai.ExecuteAI(ctx, event.UserID, event.RoomID)
	if err != nil {
		fmt.Printf("鎵ц AI 澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v\n", event.UserID, event.RoomID, err)
		return nil
	}

	if reply.Message.Content != "" {
		h.rooms.BroadcastToRoom(event.RoomID, reply)
	}

	return nil
}
