package websocket

import (
	"testing"
)

func TestNewMessageEvent(t *testing.T) {
	event := NewMessageEvent("topic1", "room1", "user1", "hello")

	if event.Type != EventTypeMessage {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeMessage)
	}
	if event.Msg.Topic != "topic1" {
		t.Errorf("Msg.Topic = %s, want topic1", event.Msg.Topic)
	}
	if string(event.Msg.Key) != "room1" {
		t.Errorf("Msg.Key = %s, want room1", string(event.Msg.Key))
	}
	data, ok := event.Data.(*MessageEvent)
	if !ok {
		t.Fatalf("Data type = %T, want *MessageEvent", event.Data)
	}
	if data.RoomID != "room1" || data.SenderID != "user1" || data.Content != "hello" {
		t.Errorf("Data = %+v, want room1/user1/hello", data)
	}
}

func TestNewUnreadEvent(t *testing.T) {
	event := NewUnreadEvent("topic2", "room2", "user2")

	if event.Type != EventTypeUnread {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeUnread)
	}
	if event.Msg.Topic != "topic2" {
		t.Errorf("Msg.Topic = %s, want topic2", event.Msg.Topic)
	}
	if string(event.Msg.Key) != "room2" {
		t.Errorf("Msg.Key = %s, want room2", string(event.Msg.Key))
	}
	data, ok := event.Data.(*UnreadEvent)
	if !ok {
		t.Fatalf("Data type = %T, want *UnreadEvent", event.Data)
	}
	if data.RoomID != "room2" || data.UserID != "user2" {
		t.Errorf("Data = %+v, want room2/user2", data)
	}
}

func TestNewRoomEvent(t *testing.T) {
	event := NewRoomEvent("topic3", "join", "room3", "user3")

	if event.Type != EventTypeRoom {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeRoom)
	}
	if event.Msg.Topic != "topic3" {
		t.Errorf("Msg.Topic = %s, want topic3", event.Msg.Topic)
	}
	if string(event.Msg.Key) != "room3" {
		t.Errorf("Msg.Key = %s, want room3", string(event.Msg.Key))
	}
	data, ok := event.Data.(*RoomEvent)
	if !ok {
		t.Fatalf("Data type = %T, want *RoomEvent", event.Data)
	}
	if data.Action != "join" || data.RoomID != "room3" || data.UserID != "user3" {
		t.Errorf("Data = %+v, want join/room3/user3", data)
	}
}

func TestNewAIChatEvent(t *testing.T) {
	event := NewAIChatEvent("topic4", "room4", "user4", "ai content")

	if event.Type != EventTypeAIChat {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeAIChat)
	}
	if event.Msg.Topic != "topic4" {
		t.Errorf("Msg.Topic = %s, want topic4", event.Msg.Topic)
	}
	if string(event.Msg.Key) != "room4" {
		t.Errorf("Msg.Key = %s, want room4", string(event.Msg.Key))
	}
	data, ok := event.Data.(*AIChatEvent)
	if !ok {
		t.Fatalf("Data type = %T, want *AIChatEvent", event.Data)
	}
	if data.RoomID != "room4" || data.UserID != "user4" || data.Content != "ai content" {
		t.Errorf("Data = %+v, want room4/user4/ai content", data)
	}
}
