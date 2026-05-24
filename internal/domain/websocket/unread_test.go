package websocket

import (
	"context"
	"errors"
	"testing"
)

func TestHandleGetUnread_NotMember(t *testing.T) {
	cache := &mockMessageCache{unreadCount: 5}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	client := newTestClient("u1") // 不在任何房间
	mm.HandleGetUnread(context.Background(), client, "r1")

	// 非成员不应读 cache 或发 Kafka
	if len(writer.events) != 0 {
		t.Error("should not send Kafka event for non-member")
	}
}

func TestHandleGetUnread_UnreadCountZero(t *testing.T) {
	cache := &mockMessageCache{unreadCount: 0}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	client := newTestClient("u1", "r1")
	mm.HandleGetUnread(context.Background(), client, "r1")

	// unread=0 时不应发历史消息到 client.Send
	select {
	case <-client.Send:
		t.Error("should not send messages to client when unread=0")
	default:
	}

	// 但会发 UnreadEvent 到 Kafka
	if len(writer.events) != 1 {
		t.Fatalf("expected 1 Kafka event, got %d", len(writer.events))
	}
	if writer.events[0].Type != EventTypeUnread {
		t.Errorf("event type = %s, want %s", writer.events[0].Type, EventTypeUnread)
	}
}

func TestHandleGetRead_UnreadCountGreaterThanZero(t *testing.T) {
	msgs := []CacheMessage{
		{ID: "m1", Context: "hello"},
		{ID: "m2", Context: "world"},
	}
	cache := &mockMessageCache{unreadCount: 2, messages: msgs}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	client := newTestClient("u1", "r1")
	mm.HandleGetUnread(context.Background(), client, "r1")

	// 应该发送 2 条消息到 client.Send
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

	// 也会发 UnreadEvent
	if len(writer.events) != 1 || writer.events[0].Type != EventTypeUnread {
		t.Error("expected UnreadEvent in Kafka")
	}
}

func TestHandleGetUnread_GetUnreadCountFails(t *testing.T) {
	cache := &mockMessageCache{unreadCountErr: errors.New("cache error")}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	client := newTestClient("u1", "r1")
	mm.HandleGetUnread(context.Background(), client, "r1")

	// 错误时不应发消息或事件
	select {
	case <-client.Send:
		t.Error("should not send messages on error")
	default:
	}
	if len(writer.events) != 0 {
		t.Error("should not send Kafka event on error")
	}
}

func TestHandleGetUnread_GetMessagesFails(t *testing.T) {
	cache := &mockMessageCache{unreadCount: 2, messagesErr: errors.New("cache error")}
	repo := &mockMessageRepo{}
	roomRepo := &mockRoomRepo{}
	writer := &mockMessageWriter{}
	rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
	mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

	client := newTestClient("u1", "r1")
	mm.HandleGetUnread(context.Background(), client, "r1")

	select {
	case <-client.Send:
		t.Error("should not send messages on GetMessages error")
	default:
	}
}

func TestUpdateUnreadCount_SkipsSender(t *testing.T) {
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
}

func TestUpdateUnreadCount_GetChatRoomUsersFails(t *testing.T) {
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
}
