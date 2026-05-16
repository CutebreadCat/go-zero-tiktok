package websocket

import (
	"context"
	"log"

	"github.com/sashabaranov/go-openai"
)

type AIChatMessage interface {
	// GetContent 获取消息内容
	Run(ctx context.Context, messages []openai.ChatCompletionMessage, jwchid, jwchpassword string) (string, error)
	ParseMessageToOpenAIList(ctx context.Context, msg []CacheMessage) []openai.ChatCompletionMessage
}
type AiCacheMessage interface {
	IncrAIMessage(ctx context.Context, userID, roomID string) error
	GetAIMessageCount(ctx context.Context, userID, roomID string) (int64, error)
	ClearAIMessage(ctx context.Context, userID, roomID string) error
	GetMessages(ctx context.Context, userID, roomID string, count int64) ([]CacheMessage, error)
}

type AIdal interface {
	GetUserJwchInfo(ctx context.Context, userID string) (string, string, error)
}

type AIChat struct {
	Agent AIChatMessage
	Cache AiCacheMessage
	Dal   AIdal
}

func NewAIChat(agent AIChatMessage, cache AiCacheMessage, dal AIdal) *AIChat {
	return &AIChat{
		Agent: agent,
		Cache: cache,
		Dal:   dal,
	}
}
func (a *AIChat) MakeAiContent(ctx context.Context, userid, roomID string) (Message, error) {
	//这里就不采用历史消息的查询还是采用实时转发,就行然后实时转发的话每达到五条进行一次调用,如果后续支持历史消息或者定时器再说.
	var jwchid, jwchpassword string
	var err error
	if jwchid, jwchpassword, err = a.Dal.GetUserJwchInfo(ctx, userid); err != nil {
		log.Printf("Failed to get Jwch info for user %s: %v", userid, err)
		return Message{}, err
	}
	var count int64

	if count, err = a.Cache.GetAIMessageCount(ctx, userid, roomID); err != nil {
		log.Printf("Failed to get AI message count for user %s in room %s: %v", userid, roomID, err)
		return Message{}, err
	}
	if count < 5 {
		log.Printf("AI message count for user %s in room %s is less than 5, current count: %d", userid, roomID, count)
		return Message{}, nil
	}
	// 获取消息列表
	messages, err := a.Cache.GetMessages(ctx, userid, roomID, count)
	if err != nil {
		log.Printf("Failed to get messages for user %s in room %s: %v", userid, roomID, err)
		return Message{}, err
	}
	// 将消息列表转换为 OpenAI API 需要的格式
	openAIMessages := a.Agent.ParseMessageToOpenAIList(ctx, messages)
	// 调用 Agent 获取 AI 回复
	reply, err := a.Agent.Run(ctx, openAIMessages, jwchid, jwchpassword)
	if err != nil {
		log.Printf("Failed to get AI reply for user %s in room %s: %v", userid, roomID, err)
		return Message{}, err
	}
	var aiMessage Message
	aiMessage.Message.RoomID = roomID
	aiMessage.Message.SenderID = "AI"
	aiMessage.Message.Content = reply
	aiMessage.Typek = "ai_reply"
	// 清除 AI 消息计数
	if err = a.Cache.ClearAIMessage(ctx, userid, roomID); err != nil {
		log.Printf("Failed to clear AI messages for user %s in room %s: %v", userid, roomID, err)
		return Message{}, err
	}

	return aiMessage, nil
}
