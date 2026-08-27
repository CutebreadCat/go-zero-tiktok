package messagestable

import "time"

// Message 消息数据库模型。
type Message struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;type:bigint;column:id"`
	MessageID       int64      `gorm:"type:bigint;column:message_id;not null;uniqueIndex:uk_message_id"`
	ReceiverID      int64      `gorm:"type:bigint;column:receiver_id;not null;index:idx_receiver_read_created,priority:1"`
	Type            string     `gorm:"type:varchar(16);column:type;not null"`
	Title           string     `gorm:"type:varchar(128);column:title;not null"`
	Content         string     `gorm:"type:varchar(1024);column:content;not null"`
	EventID         string     `gorm:"type:varchar(64);column:event_id;uniqueIndex:uk_receiver_event,priority:2"`
	SenderID        int64      `gorm:"type:bigint;column:sender_id;not null;default:0"`
	SenderNickname  string     `gorm:"type:varchar(64);column:sender_nickname;not null;default:''"`
	SenderAvatarURL string     `gorm:"type:varchar(512);column:sender_avatar_url;not null;default:''"`
	TargetID        int64      `gorm:"type:bigint;column:target_id;not null;default:0"`
	TargetType      string     `gorm:"type:varchar(32);column:target_type;not null;default:''"`
	IsRead          int8       `gorm:"type:tinyint;column:is_read;not null;default:0;index:idx_receiver_read_created,priority:2"`
	CreatedAt       time.Time  `gorm:"type:datetime(3);column:created_at;not null;index:idx_receiver_read_created,priority:3"`
	ReadAt          *time.Time `gorm:"type:datetime(3);column:read_at"`
}

// TableName 返回表名。
func (Message) TableName() string {
	return "messages"
}
