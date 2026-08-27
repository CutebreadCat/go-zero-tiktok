package domainmessage

import "context"

// Repository 定义消息领域需要的持久化能力。
type Repository interface {
	// Create 保存新消息；同一接收人同一事件重复写入时返回既有消息。
	Create(ctx context.Context, message *Message, idempotencyKey string) (*Message, bool, error)
	// ListByUser 按创建时间倒序读取当前接收人的消息列表。
	ListByUser(ctx context.Context, receiverID int64, messageType string, cursor *Cursor, limit int) ([]*Message, error)
	// CountUnread 统计当前接收人未读消息数。
	CountUnread(ctx context.Context, receiverID int64) (int64, error)
	// MarkRead 将当前接收人的指定消息标记为已读；空列表表示当前用户全部消息。
	MarkRead(ctx context.Context, receiverID int64, messageIDs []int64) (int64, error)
}
