package websocket

import (
	"testing"

	mqcontract "go_zero-tiktok/internal/shared/mq"
)

func TestNewEvents(t *testing.T) {
	tests := []struct {
		name      string
		buildFn   func() *mqcontract.Event
		wantType  string
		checkData func(t *testing.T, data interface{})
	}{
		{
			name: "NewMessageEvent",
			buildFn: func() *mqcontract.Event {
				return NewMessageEvent("topic1", "room1", "user1", "hello")
			},
			wantType: EventTypeMessage,
			checkData: func(t *testing.T, data interface{}) {
				d, ok := data.(*MessageEvent)
				if !ok {
					t.Fatalf("Data type = %T, want *MessageEvent", data)
				}
				if d.RoomID != "room1" || d.SenderID != "user1" || d.Content != "hello" {
					t.Errorf("Data = %+v, want room1/user1/hello", d)
				}
			},
		},
		{
			name: "NewUnreadEvent",
			buildFn: func() *mqcontract.Event {
				return NewUnreadEvent("topic2", "room2", "user2")
			},
			wantType: EventTypeUnread,
			checkData: func(t *testing.T, data interface{}) {
				d, ok := data.(*UnreadEvent)
				if !ok {
					t.Fatalf("Data type = %T, want *UnreadEvent", data)
				}
				if d.RoomID != "room2" || d.UserID != "user2" {
					t.Errorf("Data = %+v, want room2/user2", d)
				}
			},
		},
		{
			name: "NewRoomEvent",
			buildFn: func() *mqcontract.Event {
				return NewRoomEvent("topic3", "join", "room3", "user3")
			},
			wantType: EventTypeRoom,
			checkData: func(t *testing.T, data interface{}) {
				d, ok := data.(*RoomEvent)
				if !ok {
					t.Fatalf("Data type = %T, want *RoomEvent", data)
				}
				if d.Action != "join" || d.RoomID != "room3" || d.UserID != "user3" {
					t.Errorf("Data = %+v, want join/room3/user3", d)
				}
			},
		},
		{
			name: "NewAIChatEvent",
			buildFn: func() *mqcontract.Event {
				return NewAIChatEvent("topic4", "room4", "user4", "ai content")
			},
			wantType: EventTypeAIChat,
			checkData: func(t *testing.T, data interface{}) {
				d, ok := data.(*AIChatEvent)
				if !ok {
					t.Fatalf("Data type = %T, want *AIChatEvent", data)
				}
				if d.RoomID != "room4" || d.UserID != "user4" || d.Content != "ai content" {
					t.Errorf("Data = %+v, want room4/user4/ai content", d)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.buildFn()

			if event.Type != tt.wantType {
				t.Errorf("Type = %s, want %s", event.Type, tt.wantType)
			}

			// 验证 Msg.Topic 和 Msg.Key
			switch tt.wantType {
			case EventTypeMessage:
				if event.Msg.Topic != "topic1" {
					t.Errorf("Msg.Topic = %s, want topic1", event.Msg.Topic)
				}
				if string(event.Msg.Key) != "room1" {
					t.Errorf("Msg.Key = %s, want room1", string(event.Msg.Key))
				}
			case EventTypeUnread:
				if event.Msg.Topic != "topic2" {
					t.Errorf("Msg.Topic = %s, want topic2", event.Msg.Topic)
				}
				if string(event.Msg.Key) != "room2" {
					t.Errorf("Msg.Key = %s, want room2", string(event.Msg.Key))
				}
			case EventTypeRoom:
				if event.Msg.Topic != "topic3" {
					t.Errorf("Msg.Topic = %s, want topic3", event.Msg.Topic)
				}
				if string(event.Msg.Key) != "room3" {
					t.Errorf("Msg.Key = %s, want room3", string(event.Msg.Key))
				}
			case EventTypeAIChat:
				if event.Msg.Topic != "topic4" {
					t.Errorf("Msg.Topic = %s, want topic4", event.Msg.Topic)
				}
				if string(event.Msg.Key) != "room4" {
					t.Errorf("Msg.Key = %s, want room4", string(event.Msg.Key))
				}
			}

			tt.checkData(t, event.Data)
		})
	}
}
