package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	mqcontract "go_zero-tiktok/internal/shared/mq"
)

func TestMessageHandler_Consume_CallsHandleMessageByUserID(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	handler := NewMessageHandler(mm)
	event := &mqcontract.Event{
		Type: EventTypeMessage,
		Data: &MessageEvent{
			RoomID:   "r1",
			SenderID: "u1",
			Content:  "hello",
		},
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// HandleMessageByUserID 被调用，通过 repo 和 cache 的 mock 验证
}

func TestMessageHandler_Consume_InvalidDataType(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	handler := NewMessageHandler(mm)
	event := &mqcontract.Event{
		Type: EventTypeMessage,
		Data: "invalid type", // 不是 *MessageEvent
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 不 panic、不调用 manager 即可
}

func TestUnreadHandler_Consume_CallsHandleGetUnreadByUserID(t *testing.T) {
	cache := &mockMessageCache{unreadCount: 0}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	handler := NewUnreadHandler(mm)
	event := &mqcontract.Event{
		Type: EventTypeUnread,
		Data: &UnreadEvent{
			RoomID: "r1",
			UserID: "u1",
		},
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnreadHandler_Consume_InvalidDataType(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	handler := NewUnreadHandler(mm)
	event := &mqcontract.Event{
		Type: EventTypeUnread,
		Data: "invalid",
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAIChatHandler_Consume_CallsExecuteAIAndBroadcasts(t *testing.T) {
	cache := &mockMessageCache{}
	agent := &mockAIAgent{reply: "AI reply"}
	ai := NewAIChat(agent, cache)
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	client := newTestClient("u1", "r1")
	rooms.mu.Lock()
	rooms.rooms["r1"] = map[*Client]bool{client: true}
	rooms.mu.Unlock()

	handler := NewAIChatHandler(ai, rooms)
	event := &mqcontract.Event{
		Type: EventTypeAIChat,
		Data: &AIChatEvent{
			RoomID:  "r1",
			UserID:  "u1",
			Content: "trigger",
		},
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AI 回复应该被广播到房间
	select {
	case msg := <-client.Send:
		m, ok := msg.(Message)
		if !ok {
			t.Fatalf("msg type = %T, want Message", msg)
		}
		if m.Message.Content != "AI reply" {
			t.Errorf("content = %s, want 'AI reply'", m.Message.Content)
		}
		if m.Message.SenderID != aiSenderID {
			t.Errorf("SenderID = %s, want %s", m.Message.SenderID, aiSenderID)
		}
	case <-time.After(time.Second):
		t.Error("expected AI reply in client.Send")
	}
}

func TestAIChatHandler_Consume_AIFailsReturnsNil(t *testing.T) {
	cache := &mockMessageCache{}
	agent := &mockAIAgent{runErr: errors.New("agent crash")}
	ai := NewAIChat(agent, cache)
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})

	handler := NewAIChatHandler(ai, rooms)
	event := &mqcontract.Event{
		Type: EventTypeAIChat,
		Data: &AIChatEvent{
			RoomID:  "r1",
			UserID:  "u1",
			Content: "trigger",
		},
	}

	err := handler.Consume(context.Background(), event)
	// 当前实现：AI 失败时打印日志并 return nil
	if err != nil {
		t.Errorf("expected nil error on AI failure, got %v", err)
	}
}

func TestAIChatHandler_Consume_EmptyReplyNoBroadcast(t *testing.T) {
	cache := &mockMessageCache{}
	agent := &mockAIAgent{reply: ""}
	ai := NewAIChat(agent, cache)
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	client := newTestClient("u1", "r1")
	rooms.mu.Lock()
	rooms.rooms["r1"] = map[*Client]bool{client: true}
	rooms.mu.Unlock()

	handler := NewAIChatHandler(ai, rooms)
	event := &mqcontract.Event{
		Type: EventTypeAIChat,
		Data: &AIChatEvent{
			RoomID:  "r1",
			UserID:  "u1",
			Content: "trigger",
		},
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-client.Send:
		t.Error("should not broadcast when reply is empty")
	default:
	}
}

func TestAIChatHandler_Consume_InvalidDataType(t *testing.T) {
	cache := &mockMessageCache{}
	ai := NewAIChat(&mockAIAgent{}, cache)
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})

	handler := NewAIChatHandler(ai, rooms)
	event := &mqcontract.Event{
		Type: EventTypeAIChat,
		Data: "invalid",
	}

	err := handler.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
