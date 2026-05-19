package websocket

import (
	"context"

	"log"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/sashabaranov/go-openai"
)

type AIChatMessage interface {
	Run(ctx context.Context, userId string, messages []openai.ChatCompletionMessage) (string, error)
	ParseMessageToOpenAIList(ctx context.Context, msg []CacheMessage) []openai.ChatCompletionMessage
}

type AiCacheMessage interface {
	IncrAIMessage(ctx context.Context, userID, roomID string) error
	GetAIMessageCount(ctx context.Context, userID, roomID string) (int64, error)
	ClearAIMessage(ctx context.Context, userID, roomID string) error
	AddAIMessage(ctx context.Context, userID, roomID string, message *types.MessageChat) (string, error)
	GetAIMessages(ctx context.Context, userID, roomID string, count int64) ([]CacheMessage, error)
	ClearAIStream(ctx context.Context, userID, roomID string) error
}

type AIChat struct {
	Agent AIChatMessage
	Cache AiCacheMessage
}

func NewAIChat(agent AIChatMessage, cache AiCacheMessage) *AIChat {
	return &AIChat{
		Agent: agent,
		Cache: cache,
	}
}

func (a *AIChat) CheckAndEnqueue(ctx context.Context, userID, roomID string, msg *types.MessageChat) (bool, error) {
	if _, err := a.Cache.AddAIMessage(ctx, userID, roomID, msg); err != nil {
		return false, xerr.Wrap(err, "AIChat.CheckAndEnqueue.AddAIMessage")
	}

	if err := a.Cache.IncrAIMessage(ctx, userID, roomID); err != nil {
		return false, xerr.Wrap(err, "AIChat.CheckAndEnqueue.IncrAIMessage")
	}

	count, err := a.Cache.GetAIMessageCount(ctx, userID, roomID)
	if err != nil {
		return false, xerr.Wrap(err, "AIChat.CheckAndEnqueue.GetAIMessageCount")
	}

	if count < 2 {
		log.Printf("用户 %s 在房间 %s 的 AI 消息计数为 %d，等待达到 2", userID, roomID, count)
		return false, nil
	}

	return true, nil
}

func (a *AIChat) ExecuteAI(ctx context.Context, userID, roomID string) (Message, error) {
	defer func() {
		if err := a.Cache.ClearAIMessage(ctx, userID, roomID); err != nil {
			log.Printf("清除 AI 消息计数失败 (用户 %s, 房间 %s): %v", userID, roomID, err)
		}
		if err := a.Cache.ClearAIStream(ctx, userID, roomID); err != nil {
			log.Printf("清除 AI 流失败 (用户 %s, 房间 %s): %v", userID, roomID, err)
		}
	}()

	messages, err := a.Cache.GetAIMessages(ctx, userID, roomID, 2)
	if err != nil {
		return Message{}, xerr.Wrap(err, "AIChat.ExecuteAI.GetAIMessages")
	}

	openAIMessages := a.Agent.ParseMessageToOpenAIList(ctx, messages)
	reply, err := a.Agent.Run(ctx, userID, openAIMessages)
	if err != nil {
		return Message{}, xerr.Wrap(err, "AIChat.ExecuteAI.Run")
	}

	return Message{
		Message: types.MessageChat{
			RoomID:   roomID,
			SenderID: "AI",
			Content:  reply,
		},
		Typek: "ai_reply",
	}, nil
}
