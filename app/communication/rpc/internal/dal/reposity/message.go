package reposity

import (
	"context"

	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
	messagestable "go_zero-tiktok/app/communication/rpc/internal/dal/tables/messages"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

// MessageRepo 消息仓库实现。
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建消息仓库。
func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// Create 保存消息，幂等冲突时返回已有消息。
func (r *MessageRepo) Create(ctx context.Context, message *domainmessage.Message, idempotencyKey string) (*domainmessage.Message, bool, error) {
	record, inserted, err := messagestable.Create(ctx, r.db, message, idempotencyKey)
	if err != nil {
		return nil, false, pkgerrors.WithMessage(err, "MessageRepo.Create")
	}
	return record.ToDomain(), inserted, nil
}

// ListByUser 查询消息列表。
func (r *MessageRepo) ListByUser(ctx context.Context, receiverID int64, messageType string, cursor *domainmessage.Cursor, limit int) ([]*domainmessage.Message, error) {
	records, err := messagestable.ListByUser(ctx, r.db, receiverID, messageType, cursor, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "MessageRepo.ListByUser")
	}

	result := make([]*domainmessage.Message, 0, len(records))
	for _, record := range records {
		result = append(result, record.ToDomain())
	}
	return result, nil
}

// CountUnread 统计未读数。
func (r *MessageRepo) CountUnread(ctx context.Context, receiverID int64) (int64, error) {
	count, err := messagestable.CountUnread(ctx, r.db, receiverID)
	if err != nil {
		return 0, pkgerrors.WithMessage(err, "MessageRepo.CountUnread")
	}
	return count, nil
}

// MarkRead 标记已读。
func (r *MessageRepo) MarkRead(ctx context.Context, receiverID int64, messageIDs []int64) (int64, error) {
	count, err := messagestable.MarkRead(ctx, r.db, receiverID, messageIDs)
	if err != nil {
		return 0, pkgerrors.WithMessage(err, "MessageRepo.MarkRead")
	}
	return count, nil
}
