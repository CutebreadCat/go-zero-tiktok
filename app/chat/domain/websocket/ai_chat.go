package websocket

import (
	"context"

	appLogger "go_zero-tiktok/Prometheus/logger"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"

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

	if count < aiTriggerMsgCount {
		appLogger.Infof("鐢ㄦ埛 %s 鍦ㄦ埧闂?%s 鐨?AI 娑堟伅璁℃暟涓?%d锛岀瓑寰呰揪鍒?%d", userID, roomID, count, aiTriggerMsgCount)
		return false, nil
	}

	return true, nil
}

func (a *AIChat) ExecuteAI(ctx context.Context, userID, roomID string) (Message, error) {
	defer func() {
		if err := a.Cache.ClearAIMessage(ctx, userID, roomID); err != nil {
			appLogger.Infof("娓呴櫎 AI 娑堟伅璁℃暟澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", userID, roomID, err)
		}
		if err := a.Cache.ClearAIStream(ctx, userID, roomID); err != nil {
			appLogger.Infof("娓呴櫎 AI 娴佸け璐?(鐢ㄦ埛 %s, 鎴块棿 %s): %v", userID, roomID, err)
		}
	}()

	messages, err := a.Cache.GetAIMessages(ctx, userID, roomID, aiTriggerMsgCount)
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
			SenderID: aiSenderID,
			Content:  reply,
		},
		Typek: MessageTypeAIReply,
	}, nil
}
