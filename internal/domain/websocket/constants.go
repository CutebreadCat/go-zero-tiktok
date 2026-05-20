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
	MessageTypeAIReply = "ai_reply"

	aiSenderID     = "AI"
	dateTimeLayout = "2006-01-02 15:04:05"

	pongWait = 60 * time.Second
)
