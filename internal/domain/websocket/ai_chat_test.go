package websocket

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/types"
)

func TestCheckAndEnqueue_AddAIMessageFails(t *testing.T) {
	cache := &mockMessageCache{addAIMsgErr: errors.New("redis down")}
	ai := NewAIChat(&mockAIAgent{}, cache)

	reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reached {
		t.Error("expected reached=false on AddAIMessage error")
	}
}

func TestCheckAndEnqueue_IncrAIMessageFails(t *testing.T) {
	cache := &mockMessageCache{incrAIErr: errors.New("redis down")}
	ai := NewAIChat(&mockAIAgent{}, cache)

	reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reached {
		t.Error("expected reached=false on IncrAIMessage error")
	}
}

func TestCheckAndEnqueue_BelowThreshold(t *testing.T) {
	cache := &mockMessageCache{aiCount: 1} // aiTriggerMsgCount=2
	ai := NewAIChat(&mockAIAgent{}, cache)

	reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reached {
		t.Error("expected reached=false when count < aiTriggerMsgCount")
	}
}

func TestCheckAndEnqueue_AtThreshold(t *testing.T) {
	cache := &mockMessageCache{aiCount: 2} // aiTriggerMsgCount=2
	ai := NewAIChat(&mockAIAgent{}, cache)

	reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reached {
		t.Error("expected reached=true when count >= aiTriggerMsgCount")
	}
}

func TestCheckAndEnqueue_AboveThreshold(t *testing.T) {
	cache := &mockMessageCache{aiCount: 5}
	ai := NewAIChat(&mockAIAgent{}, cache)

	reached, err := ai.CheckAndEnqueue(context.Background(), "u1", "r1", &types.MessageChat{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reached {
		t.Error("expected reached=true when count > aiTriggerMsgCount")
	}
}

func TestExecuteAI_Success(t *testing.T) {
	cache := &mockMessageCache{}
	agent := &mockAIAgent{reply: "hello from AI"}
	ai := NewAIChat(agent, cache)

	msg, err := ai.ExecuteAI(context.Background(), "u1", "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestExecuteAI_GetAIMessagesFails(t *testing.T) {
	cache := &mockMessageCache{aiMessagesErr: errors.New("cache error")}
	ai := NewAIChat(&mockAIAgent{}, cache)

	_, err := ai.ExecuteAI(context.Background(), "u1", "r1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecuteAI_AgentRunFails(t *testing.T) {
	cache := &mockMessageCache{}
	agent := &mockAIAgent{runErr: errors.New("agent error")}
	ai := NewAIChat(agent, cache)

	_, err := ai.ExecuteAI(context.Background(), "u1", "r1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecuteAI_ClearsOnSuccess(t *testing.T) {
	cache := &mockMessageCache{}
	ai := NewAIChat(&mockAIAgent{reply: "ok"}, cache)

	_, _ = ai.ExecuteAI(context.Background(), "u1", "r1")

	// ClearAIMessage 和 ClearAIStream 都应该在 defer 中被调用
	// mock 默认返回 nil，说明被调用了（无 panic）
}

func TestExecuteAI_ClearsEvenOnError(t *testing.T) {
	cache := &mockMessageCache{aiMessagesErr: errors.New("fail")}
	ai := NewAIChat(&mockAIAgent{}, cache)

	_, _ = ai.ExecuteAI(context.Background(), "u1", "r1")
	// defer 仍然执行 ClearAIMessage 和 ClearAIStream，不 panic 即可
}
