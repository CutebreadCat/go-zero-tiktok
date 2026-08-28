package applicationmessage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
)

const defaultLimit = 20

// Service 消息应用服务。
type Service struct {
	repo domainmessage.Repository
}

// New 创建消息应用服务。
func New(repo domainmessage.Repository) *Service {
	return &Service{repo: repo}
}

// CreateResult 创建消息结果。
type CreateResult struct {
	Message *domainmessage.Message
	Created bool
}

// ListResult 消息列表结果。
type ListResult struct {
	Items      []*domainmessage.Message
	NextCursor string
	HasMore    bool
}

// Create 创建消息。
func (s *Service) Create(ctx context.Context, receiverID int64, messageType, title, content, eventID, idempotencyKey string, senderID int64, senderNickname, senderAvatarURL string, targetID int64, targetType string) (*CreateResult, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > domainmessage.MaxIdempotencyKeyLength {
		return nil, domainmessage.ErrIdempotencyKeyTooLong
	}

	// 不给自己发消息
	if senderID > 0 && senderID == receiverID {
		return &CreateResult{Message: nil, Created: false}, nil
	}

	message, err := domainmessage.New(receiverID, messageType, title, content, eventID)
	if err != nil {
		return nil, err
	}
	message.WithSender(senderID, senderNickname, senderAvatarURL).WithTarget(targetID, targetType)

	if _, err := message.AssignID(); err != nil {
		return nil, err
	}

	created, inserted, err := s.repo.Create(ctx, message, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &CreateResult{Message: created, Created: inserted}, nil
}

// List 查询消息列表。
func (s *Service) List(ctx context.Context, receiverID int64, messageType, cursor string, limit int) (*ListResult, error) {
	if receiverID <= 0 {
		return nil, domainmessage.ErrInvalidReceiverID
	}

	messageType = strings.TrimSpace(messageType)
	if messageType != "" {
		var err error
		messageType, err = domainmessage.NormalizeType(messageType)
		if err != nil {
			return nil, err
		}
	}

	parsedCursor, err := parseCursor(cursor)
	if err != nil {
		return nil, err
	}
	limit = normalizeLimit(limit)

	items, err := s.repo.ListByUser(ctx, receiverID, messageType, parsedCursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	nextCursor := ""
	if len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeCursor(&domainmessage.Cursor{
			CreatedAt: last.CreatedAt,
			MessageID: last.ID,
		})
	}

	return &ListResult{Items: items, NextCursor: nextCursor, HasMore: hasMore}, nil
}

// CountUnread 查询未读数。
func (s *Service) CountUnread(ctx context.Context, receiverID int64) (int64, error) {
	if receiverID <= 0 {
		return 0, domainmessage.ErrInvalidReceiverID
	}
	return s.repo.CountUnread(ctx, receiverID)
}

// MarkRead 标记已读。
func (s *Service) MarkRead(ctx context.Context, receiverID int64, messageIDs []int64) (int64, error) {
	if receiverID <= 0 {
		return 0, domainmessage.ErrInvalidReceiverID
	}

	ids := make([]int64, 0, len(messageIDs))
	seen := map[int64]struct{}{}
	for _, id := range messageIDs {
		if id <= 0 {
			return 0, domainmessage.ErrInvalidMessageID
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	return s.repo.MarkRead(ctx, receiverID, ids)
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > domainmessage.MaxLimit {
		return domainmessage.MaxLimit
	}
	return limit
}

type cursorPayload struct {
	CreatedAt string `json:"created_at"`
	MessageID int64  `json:"message_id"`
}

func parseCursor(raw string) (*domainmessage.Cursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	content, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		content, err = base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, domainmessage.ErrInvalidCursor
		}
	}

	var payload cursorPayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, domainmessage.ErrInvalidCursor
	}

	createdAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(payload.CreatedAt))
	if err != nil || payload.MessageID <= 0 {
		return nil, domainmessage.ErrInvalidCursor
	}

	return &domainmessage.Cursor{CreatedAt: createdAt, MessageID: payload.MessageID}, nil
}

func encodeCursor(cursor *domainmessage.Cursor) string {
	if cursor == nil || cursor.MessageID <= 0 || cursor.CreatedAt.IsZero() {
		return ""
	}

	content, err := json.Marshal(cursorPayload{
		CreatedAt: cursor.CreatedAt.UTC().Format(time.RFC3339Nano),
		MessageID: cursor.MessageID,
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(content)
}
