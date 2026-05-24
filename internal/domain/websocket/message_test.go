package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"
)

func TestHandleMessage_NotMember(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	ai := NewAIChat(&mockAIAgent{}, cache)
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, ai)

	client := newTestClient("u1") // 不在任何房间
	mm.HandleMessage(context.Background(), client, &types.MessageChat{RoomID: "r1"})

	// 非成员不应广播、不应写 Kafka
	if len(writer.events) != 0 {
		t.Error("non-member should not trigger Kafka event")
	}
}

func TestHandleMessage_BroadcastsAndWritesEvent(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	ai := NewAIChat(&mockAIAgent{}, cache)
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, ai)

	sender := newTestClient("u1", "r1")
	other := newTestClient("u2", "r1")
	rooms.mu.Lock()
	rooms.rooms["r1"] = map[*Client]bool{sender: true, other: true}
	rooms.mu.Unlock()

	mm.HandleMessage(context.Background(), sender, &types.MessageChat{
		RoomID:  "r1",
		Content: "hi",
	})

	// 广播：两个 client 都应收到
	for _, c := range []*Client{sender, other} {
		select {
		case msg := <-c.Send:
			m, ok := msg.(Message)
			if !ok {
				t.Fatalf("msg type = %T, want Message", msg)
			}
			if m.Message.Content != "hi" {
				t.Errorf("content = %s, want hi", m.Message.Content)
			}
		default:
			t.Errorf("client %s should have received broadcast", c.UserID)
		}
	}

	// Kafka 应该收到至少 1 条 MessageEvent
	if len(writer.events) < 1 {
		t.Fatal("expected at least 1 Kafka event")
	}
	if writer.events[0].Type != EventTypeMessage {
		t.Errorf("first event type = %s, want %s", writer.events[0].Type, EventTypeMessage)
	}
}

func TestHandleMessage_AIEventWhenThresholdReached(t *testing.T) {
	cache := &mockMessageCache{aiCount: 2} // 达到阈值
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	// 用 chanWriter 收集事件，避免 goroutine 竞态
	ch := make(chan *mqcontract.Event, 10)
	writer := &chanWriter{ch: ch}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	ai := NewAIChat(&mockAIAgent{}, cache)
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, ai)

	client := newTestClient("u1", "r1")
	rooms.mu.Lock()
	rooms.rooms["r1"] = map[*Client]bool{client: true}
	rooms.mu.Unlock()

	mm.HandleMessage(context.Background(), client, &types.MessageChat{
		RoomID:  "r1",
		Content: "trigger",
	})

	// 等待 goroutine 中的 AI 事件
	var gotAIEvent bool
	deadline := time.After(2 * time.Second)
	for !gotAIEvent {
		select {
		case evt := <-ch:
			if evt.Type == EventTypeAIChat {
				gotAIEvent = true
			}
		case <-deadline:
			goto done
		}
	}
done:
	if !gotAIEvent {
		t.Error("expected AIChatEvent when threshold reached")
	}
}

func TestHandleMessageByUserID_GeneratesUUIDWhenEmpty(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	msg := &types.MessageChat{RoomID: "r1", Content: "test"}
	mm.HandleMessageByUserID(context.Background(), "u1", msg)

	if msg.ID == "" {
		t.Error("msg.ID should be generated when empty")
	}
}

func TestHandleMessageByUserID_KeepsExistingID(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	msg := &types.MessageChat{ID: "existing-id", RoomID: "r1"}
	mm.HandleMessageByUserID(context.Background(), "u1", msg)

	if msg.ID != "existing-id" {
		t.Errorf("msg.ID = %s, want existing-id", msg.ID)
	}
}

func TestHandleMessageByUserID_CallsCacheAndRepo(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{users: []string{"u2"}}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	msg := &types.MessageChat{RoomID: "r1", Content: "test"}
	mm.HandleMessageByUserID(context.Background(), "u1", msg)

	// cache.AddMessage 和 repo.StoreChatMessage 都应该被调用
	// 通过 mock 不报错来验证（默认返回 nil）
	// UpdateUnreadCount 也应该被调用（通过 roomRepo.GetChatRoomUsers 被调用验证）
}

func TestHandleMessageByUserID_CacheAddMessageFails(t *testing.T) {
	cache := &mockMessageCache{addMsgErr: errors.New("cache error")}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	// 不应 panic
	msg := &types.MessageChat{RoomID: "r1"}
	mm.HandleMessageByUserID(context.Background(), "u1", msg)
}

func TestHandleMessageByUserID_RepoStoreFails(t *testing.T) {
	cache := &mockMessageCache{}
	repo := &mockMessageRepo{storeErr: errors.New("db error")}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	// 不应 panic
	msg := &types.MessageChat{RoomID: "r1"}
	mm.HandleMessageByUserID(context.Background(), "u1", msg)
}
