package domainmessage

import (
	"strings"
	"time"

	myutils "go_zero-tiktok/pkg/utils"
)

const (
	TypeLike    = "LIKE"
	TypeComment = "COMMENT"
	TypeFollow  = "FOLLOW"
	TypeSystem  = "SYSTEM"

	MaxTitleLength          = 128
	MaxContentLength        = 1024
	MaxEventIDLength        = 64
	MaxIdempotencyKeyLength = 128
	MaxLimit                = 100
	DefaultLimit            = 20
)

// Message 表示一个用户收到的站内通知。
type Message struct {
	ID              int64
	ReceiverID      int64
	Type            string
	Title           string
	Content         string
	EventID         string
	SenderID        int64
	SenderNickname  string
	SenderAvatarURL string
	TargetID        int64
	TargetType      string
	IsRead          bool
	CreatedAt       time.Time
	ReadAt          *time.Time
}

// Cursor 保存消息列表分页的排序字段。
type Cursor struct {
	CreatedAt time.Time
	MessageID int64
}

// New 创建消息领域对象，负责接收人、类型、标题、内容和事件 ID 校验。
func New(receiverID int64, messageType, title, content, eventID string) (*Message, error) {
	if receiverID <= 0 {
		return nil, ErrInvalidReceiverID
	}

	messageType, err := NormalizeType(messageType)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	eventID = strings.TrimSpace(eventID)

	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len(title) > MaxTitleLength {
		return nil, ErrTitleTooLong
	}
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > MaxContentLength {
		return nil, ErrContentTooLong
	}
	if len(eventID) > MaxEventIDLength {
		return nil, ErrEventIDTooLong
	}

	return &Message{
		ReceiverID: receiverID,
		Type:       messageType,
		Title:      title,
		Content:    content,
		EventID:    eventID,
		IsRead:     false,
	}, nil
}

// WithSender 写入触发消息的用户展示信息。
func (m *Message) WithSender(senderID int64, nickname, avatarURL string) *Message {
	if m == nil {
		return nil
	}
	if senderID > 0 {
		m.SenderID = senderID
	}
	m.SenderNickname = strings.TrimSpace(nickname)
	m.SenderAvatarURL = strings.TrimSpace(avatarURL)
	return m
}

// WithTarget 写入关联目标信息。
func (m *Message) WithTarget(targetID int64, targetType string) *Message {
	if m == nil {
		return nil
	}
	if targetID > 0 {
		m.TargetID = targetID
	}
	m.TargetType = strings.TrimSpace(targetType)
	return m
}

// AssignID 为消息分配雪花 ID。
func (m *Message) AssignID() (*Message, error) {
	if m == nil {
		return nil, ErrNilMessage
	}
	m.ID = myutils.GenerateMessageID()
	return m, nil
}

// Restore 从数据库记录恢复消息领域对象。
func Restore(id, receiverID int64, messageType, title, content, eventID string, senderID int64, senderNickname, senderAvatarURL string, targetID int64, targetType string, isRead bool, createdAt time.Time, readAt *time.Time) *Message {
	messageType, _ = NormalizeType(messageType)
	return &Message{
		ID:              id,
		ReceiverID:      receiverID,
		Type:            messageType,
		Title:           strings.TrimSpace(title),
		Content:         strings.TrimSpace(content),
		EventID:         strings.TrimSpace(eventID),
		SenderID:        senderID,
		SenderNickname:  strings.TrimSpace(senderNickname),
		SenderAvatarURL: strings.TrimSpace(senderAvatarURL),
		TargetID:        targetID,
		TargetType:      strings.TrimSpace(targetType),
		IsRead:          isRead,
		CreatedAt:       createdAt,
		ReadAt:          readAt,
	}
}

// NormalizeType 统一消息类型为大写。
func NormalizeType(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	switch value {
	case TypeLike, TypeComment, TypeFollow, TypeSystem:
		return value, nil
	default:
		return "", ErrInvalidMessageType
	}
}
