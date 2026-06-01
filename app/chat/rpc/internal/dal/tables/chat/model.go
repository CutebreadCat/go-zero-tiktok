package chat

import (
	"time"
)

// UserChat 用户聊天室数据库模型
type UserChat struct {
	UserID   string `gorm:"type:varchar(64);column:user_id"`
	RoomID   string `gorm:"primaryKey;type:varchar(64);column:room_id"`
	Leix     int32  `gorm:"default:0;type:int;column:leix"`
	RoomName string `gorm:"type:varchar(255);column:room_name"`
}

func (UserChat) TableName() string {
	return "user_chat"
}

// MessageChat 聊天消息数据库模型
type MessageChat struct {
	ID        string    `gorm:"primaryKey;type:varchar(64);column:id"`
	RoomID    string    `gorm:"not null;type:varchar(64);column:room_id"`
	SenderID  string    `gorm:"not null;type:varchar(64);column:sender_id"`
	Content   string    `gorm:"not null;type:varchar(1024);column:content"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
}

func (MessageChat) TableName() string {
	return "message_chat"
}
