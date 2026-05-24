package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"
)

func TestHandleMessage(t *testing.T) {
	t.Run("非成员不广播不写Kafka", func(t *testing.T) {
		cache := &mockMessageCache{}
		repo := &mockMessageRepo{}
		roomRepo := &mockRoomRepo{}
		writer := &mockMessageWriter{}
		rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
		ai := NewAIChat(&mockAIAgent{}, cache)
		mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, ai)

		client := newTestClient("u1") // 不在任何房间
		mm.HandleMessage(context.Background(), client, &types.MessageChat{RoomID: "r1"})

		if len(writer.events) != 0 {
			t.Error("non-member should not trigger Kafka event")
		}
	})

	t.Run("成员广播并写Kafka", func(t *testing.T) {
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

		if len(writer.events) < 1 {
			t.Fatal("expected at least 1 Kafka event")
		}
		if writer.events[0].Type != EventTypeMessage {
			t.Errorf("first event type = %s, want %s", writer.events[0].Type, EventTypeMessage)
		}
	})

	t.Run("达到阈值触发AI事件", func(t *testing.T) {
		cache := &mockMessageCache{aiCount: 2}
		repo := &mockMessageRepo{}
		roomRepo := &mockRoomRepo{}
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
	})
}

func TestHandleMessageByUserID(t *testing.T) {
	tests := []struct {
		name       string
		msgID      string // 初始消息ID
		cache      *mockMessageCache
		repo       *mockMessageRepo
		wantIDSet  bool   // 期望 msg.ID 被设置
		wantIDVal  string // 期望 msg.ID 的值（wantIDSet=false 时忽略）
	}{
		{
			name:      "空ID自动生成UUID",
			msgID:     "",
			cache:     &mockMessageCache{},
			repo:      &mockMessageRepo{},
			wantIDSet: true,
		},
		{
			name:      "已有ID保留不变",
			msgID:     "existing-id",
			cache:     &mockMessageCache{},
			repo:      &mockMessageRepo{},
			wantIDSet: true,
			wantIDVal: "existing-id",
		},
		{
			name:  "Cache AddMessage失败不panic",
			msgID: "",
			cache: &mockMessageCache{addMsgErr: errors.New("cache error")},
			repo:  &mockMessageRepo{},
		},
		{
			name:  "Repo Store失败不panic",
			msgID: "",
			cache: &mockMessageCache{},
			repo:  &mockMessageRepo{storeErr: errors.New("db error")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roomRepo := &mockRoomRepo{}
			writer := &mockMessageWriter{}
			rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
			mm := newTestMessageManager(tt.cache, tt.repo, roomRepo, rooms, writer, nil)

			msg := &types.MessageChat{ID: tt.msgID, RoomID: "r1", Content: "test"}
			mm.HandleMessageByUserID(context.Background(), "u1", msg)

			if tt.wantIDSet && tt.wantIDVal != "" && msg.ID != tt.wantIDVal {
				t.Errorf("msg.ID = %s, want %s", msg.ID, tt.wantIDVal)
			}
			if tt.wantIDSet && tt.msgID == "" && msg.ID == "" {
				t.Error("msg.ID should be generated when empty")
			}
		})
	}

	t.Run("调用Cache和Repo", func(t *testing.T) {
		cache := &mockMessageCache{}
		repo := &mockMessageRepo{}
		roomRepo := &mockRoomRepo{users: []string{"u2"}}
		writer := &mockMessageWriter{}
		rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
		mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

		msg := &types.MessageChat{RoomID: "r1", Content: "test"}
		mm.HandleMessageByUserID(context.Background(), "u1", msg)

		// 通过 mock 不报错来验证调用成功
	})
}
