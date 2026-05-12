package websocket

import (
	mqcontract "go_zero-tiktok/internal/shared/mq"
)

// 事件类型常量
const (
	EventTypeUnread  = "get_unread"
	EventTypeMessage = "message"
	EventTypeRoom    = "room"
)

// ============ 事件数据结构 ============

// MessageEvent 消息事件数据
type MessageEvent struct {
	RoomID   string `json:"room_id"`
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
}

// UnreadEvent 获取未读消息事件数据
type UnreadEvent struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

// RoomEvent 房间事件数据
type RoomEvent struct {
	Action string `json:"action"` // join, leave
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

// ============ 事件构造函数 ============

// NewMessageEvent 创建消息事件
func NewMessageEvent(topic string, roomID, senderID, content string) *mqcontract.Event {
	return &mqcontract.Event{
		Type: EventTypeMessage,
		Msg: &mqcontract.Message{
			Topic: topic,
			Key:   []byte(roomID),
		},
		Data: MessageEvent{
			RoomID:   roomID,
			SenderID: senderID,
			Content:  content,
		},
	}
}

// NewUnreadEvent 创建获取未读消息事件
func NewUnreadEvent(topic string, roomID, userID string) *mqcontract.Event {
	return &mqcontract.Event{
		Type: EventTypeUnread,
		Msg: &mqcontract.Message{
			Topic: topic,
			Key:   []byte(roomID),
		},
		Data: UnreadEvent{
			RoomID: roomID,
			UserID: userID,
		},
	}
}

// NewRoomEvent 创建房间事件
func NewRoomEvent(topic, action, roomID, userID string) *mqcontract.Event {
	return &mqcontract.Event{
		Type: EventTypeRoom,
		Msg: &mqcontract.Message{
			Topic: topic,
			Key:   []byte(roomID),
		},
		Data: RoomEvent{
			Action: action,
			RoomID: roomID,
			UserID: userID,
		},
	}
}
