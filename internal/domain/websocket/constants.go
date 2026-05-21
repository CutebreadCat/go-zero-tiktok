package websocket

import "time"

const (
	MessageTopic = "chat-messages"

	EventTypeUnread  = "get_unread"
	EventTypeMessage = "message"
	EventTypeRoom    = "room"
	EventTypeAIChat  = "ai_chat"

	MessageTypePing    = "ping"
	MessageTypePong    = "pong"
	MessageTypeError   = "error"
	MessageTypeAIReply = "ai_reply"

	DefaultLimitSeconds     = 1
	DefaultLimitMaxRequests = 30
	DefaultLimitKeyPrefix   = "ws_limit"

	aiSenderID          = "AI"
	aiTriggerMsgCount   = 2
	clientReadLimit     = 512
	clientContextTTL    = 24 * time.Hour
	dateTimeLayout      = "2006-01-02 15:04:05"
	pongWait            = 60 * time.Second
	rateLimitMessageKey = "message"
	typeKey             = "typek"
	wsRateLimitMessage  = "消息发送过于频繁，请稍后再试"
)
