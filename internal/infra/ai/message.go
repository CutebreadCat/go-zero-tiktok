package ai

import (
	"context"
	"go_zero-tiktok/app/chat/domain/websocket"

	"github.com/sashabaranov/go-openai"
)

func (a *Agent) ParseMessageToOpenAI(ctx context.Context, msg websocket.CacheMessage) openai.ChatCompletionMessage {
	// 将 WebSocket 消息转换为 OpenAI 消息格式
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg.Context,
	}
}
func (a *Agent) ParseMessageToOpenAIList(ctx context.Context, msg []websocket.CacheMessage) []openai.ChatCompletionMessage {
	var openaiMsgs []openai.ChatCompletionMessage
	for _, m := range msg {
		openaiMsgs = append(openaiMsgs, a.ParseMessageToOpenAI(ctx, m))
	}
	return openaiMsgs
}
