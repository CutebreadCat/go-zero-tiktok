package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	mqcontract "go_zero-tiktok/internal/shared/mq"
)

func TestMessageHandler_Consume(t *testing.T) {
	tests := []struct {
		name    string
		event   *mqcontract.Event
		wantErr bool
	}{
		{
			name: "正常消费消息",
			event: &mqcontract.Event{
				Type: EventTypeMessage,
				Data: &MessageEvent{
					RoomID:   "r1",
					SenderID: "u1",
					Content:  "hello",
				},
			},
			wantErr: false,
		},
		{
			name: "数据类型无效不报错",
			event: &mqcontract.Event{
				Type: EventTypeMessage,
				Data: "invalid type",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockMessageCache{}
			repo := &mockMessageRepo{}
			roomRepo := &mockRoomRepo{}
			writer := &mockMessageWriter{}
			rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
			mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

			handler := NewMessageHandler(mm)
			err := handler.Consume(context.Background(), tt.event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnreadHandler_Consume(t *testing.T) {
	tests := []struct {
		name    string
		event   *mqcontract.Event
		wantErr bool
	}{
		{
			name: "正常消费未读事件",
			event: &mqcontract.Event{
				Type: EventTypeUnread,
				Data: &UnreadEvent{
					RoomID: "r1",
					UserID: "u1",
				},
			},
			wantErr: false,
		},
		{
			name: "数据类型无效不报错",
			event: &mqcontract.Event{
				Type: EventTypeUnread,
				Data: "invalid",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockMessageCache{unreadCount: 0}
			repo := &mockMessageRepo{}
			roomRepo := &mockRoomRepo{}
			writer := &mockMessageWriter{}
			rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
			mm := newTestMessageManager(cache, repo, roomRepo, rooms, writer, nil)

			handler := NewUnreadHandler(mm)
			err := handler.Consume(context.Background(), tt.event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAIChatHandler_Consume(t *testing.T) {
	tests := []struct {
		name         string
		agent        *mockAIAgent
		event        *mqcontract.Event
		wantErr      bool
		wantReply    bool   // 期望客户端收到回复
		wantContent  string // 期望的回复内容
	}{
		{
			name:  "成功执行并广播",
			agent: &mockAIAgent{reply: "AI reply"},
			event: &mqcontract.Event{
				Type: EventTypeAIChat,
				Data: &AIChatEvent{
					RoomID:  "r1",
					UserID:  "u1",
					Content: "trigger",
				},
			},
			wantErr:     false,
			wantReply:   true,
			wantContent: "AI reply",
		},
		{
			name:  "AI失败返回nil",
			agent: &mockAIAgent{runErr: errors.New("agent crash")},
			event: &mqcontract.Event{
				Type: EventTypeAIChat,
				Data: &AIChatEvent{
					RoomID:  "r1",
					UserID:  "u1",
					Content: "trigger",
				},
			},
			wantErr:   false,
			wantReply: false,
		},
		{
			name:  "空回复不广播",
			agent: &mockAIAgent{reply: ""},
			event: &mqcontract.Event{
				Type: EventTypeAIChat,
				Data: &AIChatEvent{
					RoomID:  "r1",
					UserID:  "u1",
					Content: "trigger",
				},
			},
			wantErr:   false,
			wantReply: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockMessageCache{}
			ai := NewAIChat(tt.agent, cache)
			rooms := newTestRoomManager(&mockRoomRepo{}, &mockRoomCache{})
			client := newTestClient("u1", "r1")
			rooms.mu.Lock()
			rooms.rooms["r1"] = map[*Client]bool{client: true}
			rooms.mu.Unlock()

			handler := NewAIChatHandler(ai, rooms)
			err := handler.Consume(context.Background(), tt.event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantReply {
				select {
				case msg := <-client.Send:
					m, ok := msg.(Message)
					if !ok {
						t.Fatalf("msg type = %T, want Message", msg)
					}
					if m.Message.Content != tt.wantContent {
						t.Errorf("content = %s, want %s", m.Message.Content, tt.wantContent)
					}
					if m.Message.SenderID != aiSenderID {
						t.Errorf("SenderID = %s, want %s", m.Message.SenderID, aiSenderID)
					}
				case <-time.After(time.Second):
					t.Error("expected AI reply in client.Send")
				}
			} else {
				select {
				case <-client.Send:
					t.Error("should not broadcast to client")
				default:
				}
			}
		})
	}

	t.Run("数据类型无效不报错", func(t *testing.T) {
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
	})
}
