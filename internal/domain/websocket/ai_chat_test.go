package websocket

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/types"
)

func TestCheckAndEnqueue(t *testing.T) {
	tests := []struct {
		name      string
		cache     *mockMessageCache
		wantErr   bool
		wantReach bool
	}{
		{
			name:      "AddAIMessage失败",
			cache:     &mockMessageCache{addAIMsgErr: errors.New("redis down")},
			wantErr:   true,
			wantReach: false,
		},
		{
			name:      "IncrAIMessage失败",
			cache:     &mockMessageCache{incrAIErr: errors.New("redis down")},
			wantErr:   true,
			wantReach: false,
		},
		{
			name:      "低于阈值不触发",
			cache:     &mockMessageCache{aiCount: 1},
			wantErr:   false,
			wantReach: false,
		},
		{
			name:      "等于阈值触发",
			cache:     &mockMessageCache{aiCount: 2},
			wantErr:   false,
			wantReach: true,
		},
		{
			name:      "高于阈值触发",
			cache:     &mockMessageCache{aiCount: 5},
			wantErr:   false,
			wantReach: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := NewAIChat(&mockAIAgent{}, tt.cache)
			reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if reached != tt.wantReach {
				t.Errorf("reached = %v, want %v", reached, tt.wantReach)
			}
		})
	}
}

func TestExecuteAI(t *testing.T) {
	tests := []struct {
		name     string
		cache    *mockMessageCache
		agent    *mockAIAgent
		wantErr  bool
		wantMsg  bool // 是否检查返回的消息内容
	}{
		{
			name:    "成功执行",
			cache:   &mockMessageCache{},
			agent:   &mockAIAgent{reply: "hello from AI"},
			wantErr: false,
			wantMsg: true,
		},
		{
			name:    "GetAIMessages失败",
			cache:   &mockMessageCache{aiMessagesErr: errors.New("cache error")},
			agent:   &mockAIAgent{},
			wantErr: true,
			wantMsg: false,
		},
		{
			name:    "AgentRun失败",
			cache:   &mockMessageCache{},
			agent:   &mockAIAgent{runErr: errors.New("agent error")},
			wantErr: true,
			wantMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := NewAIChat(tt.agent, tt.cache)
			msg, err := ai.ExecuteAI(context.Background(), "u1", "r1")

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantMsg {
				if msg.Typek != MessageTypeAIReply {
					t.Errorf("Typek = %s, want %s", msg.Typek, MessageTypeAIReply)
				}
				if msg.Message.SenderID != aiSenderID {
					t.Errorf("SenderID = %s, want %s", msg.Message.SenderID, aiSenderID)
				}
				if msg.Message.Content != "hello from AI" {
					t.Errorf("Content = %s, want 'hello from AI'", msg.Message.Content)
				}
				if msg.Message.RoomID != "r1" {
					t.Errorf("RoomID = %s, want r1", msg.Message.RoomID)
				}
			}
		})
	}

	// 测试 defer 清理逻辑（成功和失败都不 panic）
	t.Run("成功后执行清理", func(t *testing.T) {
		cache := &mockMessageCache{}
		ai := NewAIChat(&mockAIAgent{reply: "ok"}, cache)
		_, _ = ai.ExecuteAI(context.Background(), "u1", "r1")
	})

	t.Run("失败后也执行清理", func(t *testing.T) {
		cache := &mockMessageCache{aiMessagesErr: errors.New("fail")}
		ai := NewAIChat(&mockAIAgent{}, cache)
		_, _ = ai.ExecuteAI(context.Background(), "u1", "r1")
	})
}
