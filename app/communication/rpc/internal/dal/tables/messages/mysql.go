package messagestable

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
	"go_zero-tiktok/pkg/xerr"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// Create 保存消息；幂等冲突时返回已有记录。
func Create(ctx context.Context, db *gorm.DB, message *domainmessage.Message, idempotencyKey string) (*Message, bool, error) {
	if message == nil {
		return nil, false, domainmessage.ErrNilMessage
	}

	record := &Message{
		MessageID:       message.ID,
		ReceiverID:      message.ReceiverID,
		Type:            message.Type,
		Title:           message.Title,
		Content:         message.Content,
		EventID:         message.EventID,
		SenderID:        message.SenderID,
		SenderNickname:  message.SenderNickname,
		SenderAvatarURL: message.SenderAvatarURL,
		TargetID:        message.TargetID,
		TargetType:      message.TargetType,
		IsRead:          0,
		CreatedAt:       message.CreatedAt,
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}

	err := db.WithContext(ctx).Create(record).Error
	if err == nil {
		return record, true, nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		// 幂等命中：按 event_id 或 message_id 查找已有记录
		var existing Message
		query := db.WithContext(ctx).Where("receiver_id = ?", message.ReceiverID)
		if message.EventID != "" {
			if err := query.Where("event_id = ?", message.EventID).First(&existing).Error; err == nil {
				return &existing, false, nil
			}
		}
		if err := db.WithContext(ctx).Where("message_id = ?", message.ID).First(&existing).Error; err == nil {
			return &existing, false, nil
		}
		return nil, false, xerr.Wrap(err, "MessageMySQL.Create duplicate but not found")
	}

	return nil, false, xerr.Wrap(err, "MessageMySQL.Create")
}

// ListByUser 按创建时间倒序读取接收人的消息列表。
func ListByUser(ctx context.Context, db *gorm.DB, receiverID int64, messageType string, cursor *domainmessage.Cursor, limit int) ([]Message, error) {
	query := db.WithContext(ctx).Model(&Message{}).Where("receiver_id = ?", receiverID)
	if messageType != "" {
		query = query.Where("type = ?", messageType)
	}
	if cursor != nil {
		query = query.Where(
			"created_at < ? OR (created_at = ? AND id < ?)",
			cursor.CreatedAt, cursor.CreatedAt, cursor.MessageID,
		)
	}

	var records []Message
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, xerr.Wrap(err, "MessageMySQL.ListByUser")
	}
	return records, nil
}

// CountUnread 统计未读数。
func CountUnread(ctx context.Context, db *gorm.DB, receiverID int64) (int64, error) {
	var count int64
	if err := db.WithContext(ctx).Model(&Message{}).
		Where("receiver_id = ? AND is_read = 0", receiverID).
		Count(&count).Error; err != nil {
		return 0, xerr.Wrap(err, "MessageMySQL.CountUnread")
	}
	return count, nil
}

// MarkRead 标记已读。
func MarkRead(ctx context.Context, db *gorm.DB, receiverID int64, messageIDs []int64) (int64, error) {
	query := db.WithContext(ctx).Model(&Message{}).
		Where("receiver_id = ? AND is_read = 0", receiverID)
	if len(messageIDs) > 0 {
		query = query.Where("id IN ?", messageIDs)
	}

	now := time.Now()
	result := query.Update("read_at", now).Update("is_read", 1)
	if result.Error != nil {
		return 0, xerr.Wrap(result.Error, "MessageMySQL.MarkRead")
	}
	return result.RowsAffected, nil
}

// ToDomain 将数据库模型转换为领域对象。
func (m *Message) ToDomain() *domainmessage.Message {
	if m == nil {
		return nil
	}
	return domainmessage.Restore(
		m.ID,
		m.ReceiverID,
		m.Type,
		m.Title,
		m.Content,
		m.EventID,
		m.SenderID,
		m.SenderNickname,
		m.SenderAvatarURL,
		m.TargetID,
		m.TargetType,
		m.IsRead == 1,
		m.CreatedAt,
		m.ReadAt,
	)
}

// ErrNoRows 别名，便于上层判断。
var ErrNoRows = sql.ErrNoRows
