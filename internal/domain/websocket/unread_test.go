package websocket

import (
	"context"
	"errors"
	"testing"
)

func TestHandleGetUnread(t *testing.T) {
	msgs := []CacheMessage{
		{ID: "m1", Context: "hello"},
		{ID: "m2", Context: "world"},
	}

	tests := []struct {
		name           string
		cache          *mockMessageCache
		clientRooms    []string // client 加入的房间
		wantMessages   int      // 期望发送到 client.Send 的消息数
		wantKafkaEvent bool     // 期望发送 Kafka 事件
		wantEventType  string   // 期望的 Kafka 事件类型
	}{
		{
			name:           "非成员不发消息和事件",
			cache:          &mockMessageCache{unreadCount: 5},
			clientRooms:    nil,
			wantMessages:   0,
			wantKafkaEvent: false,
		},
		{
			name:           "unread为0不发消息但发事件",
			cache:          &mockMessageCache{unreadCount: 0},
			clientRooms:    []string{"r1"},
			wantMessages:   0,
			wantKafkaEvent: true,
			wantEventType:  EventTypeUnread,
		},
		{
			name:           "unread大于0发消息和事件",
			cache:          &mockMessageCache{unreadCount: 2, messages: msgs},
			clientRooms:    []string{"r1"},
			wantMessages:   2,
			wantKafkaEvent: true,
			wantEventType:  EventTypeUnread,
		},
		{
			name:           "GetUnreadCount失败不发消息和事件",
			cache:          &mockMessageCache{unreadCountErr: errors.New("cache error")},
			clientRooms:    []string{"r1"},
			wantMessages:   0,
			wantKafkaEvent: false,
		},
		{
			name:           "GetMessages失败不发消息",
			cache:          &mockMessageCache{unreadCount: 2, messagesErr: errors.New("cache error")},
			clientRooms:    []string{"r1"},
			wantMessages:   0,
			wantKafkaEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMessageRepo{}
			roomRepo := &mockRoomRepo{}
			writer := &mockMessageWriter{}
			rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
			mm := newTestMessageManager(tt.cache, repo, roomRepo, rooms, writer, nil)

			client := newTestClient("u1", tt.clientRooms...)
			mm.HandleGetUnread(context.Background(), client, "r1")

			// 检查发送到 client.Send 的消息数
			gotMessages := 0
			for {
				select {
				case <-client.Send:
					gotMessages++
				default:
					goto done
				}
			}
		done:
			if gotMessages != tt.wantMessages {
				t.Errorf("client.Send messages = %d, want %d", gotMessages, tt.wantMessages)
			}

			// 检查 Kafka 事件
			if tt.wantKafkaEvent {
				if len(writer.events) != 1 {
					t.Fatalf("Kafka events = %d, want 1", len(writer.events))
				}
				if writer.events[0].Type != tt.wantEventType {
					t.Errorf("event type = %s, want %s", writer.events[0].Type, tt.wantEventType)
				}
			} else {
				if len(writer.events) != 0 {
					t.Errorf("Kafka events = %d, want 0", len(writer.events))
				}
			}
		})
	}

	// unread大于0时验证消息顺序
	t.Run("消息按顺序发送", func(t *testing.T) {
		cache := &mockMessageCache{unreadCount: 2, messages: msgs}
		repo := &mockMessageRepo{}
		roomRepo := &mockRoomRepo{}
		writer := &mockMessageWriter{}
		rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
		mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

		client := newTestClient("u1", "r1")
		mm.HandleGetUnread(context.Background(), client, "r1")

		for i := 0; i < 2; i++ {
			select {
			case msg := <-client.Send:
				cm, ok := msg.(CacheMessage)
				if !ok {
					t.Fatalf("msg type = %T, want CacheMessage", msg)
				}
				if cm.ID != msgs[i].ID {
					t.Errorf("msg[%d].ID = %s, want %s", i, cm.ID, msgs[i].ID)
				}
			default:
				t.Fatalf("expected message %d in client.Send", i)
			}
		}
	})
}

func TestUpdateUnreadCount(t *testing.T) {
	t.Run("跳过发送者", func(t *testing.T) {
		cache := &mockMessageCache{}
		roomRepo := &mockRoomRepo{users: []string{"u1", "u2", "u3"}}
		repo := &mockMessageRepo{}
		writer := &mockMessageWriter{}
		rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
		mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

		mm.UpdateUnreadCount(context.Background(), "u1", "r1")

		cache.mu.Lock()
		defer cache.mu.Unlock()

		if len(cache.incrUnreadCalls) != 2 {
			t.Fatalf("IncrUnread called %d times, want 2", len(cache.incrUnreadCalls))
		}
		for _, call := range cache.incrUnreadCalls {
			if call.userID == "u1" {
				t.Error("sender u1 should not have IncrUnread called")
			}
		}
	})

	t.Run("GetChatRoomUsers失败不调用IncrUnread", func(t *testing.T) {
		cache := &mockMessageCache{}
		roomRepo := &mockRoomRepo{usersErr: errors.New("db error")}
		repo := &mockMessageRepo{}
		writer := &mockMessageWriter{}
		rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
		mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

		mm.UpdateUnreadCount(context.Background(), "u1", "r1")

		cache.mu.Lock()
		defer cache.mu.Unlock()
		if len(cache.incrUnreadCalls) != 0 {
			t.Errorf("IncrUnread should not be called on error, called %d times", len(cache.incrUnreadCalls))
		}
	})
}
