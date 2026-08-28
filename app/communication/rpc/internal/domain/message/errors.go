package domainmessage

import "errors"

var (
	ErrInvalidReceiverID = errors.New("invalid receiver id")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrEmptyTitle         = errors.New("title is empty")
	ErrTitleTooLong       = errors.New("title is too long")
	ErrEmptyContent       = errors.New("content is empty")
	ErrContentTooLong     = errors.New("content is too long")
	ErrEventIDTooLong     = errors.New("event id is too long")
	ErrIdempotencyKeyTooLong = errors.New("idempotency key is too long")
	ErrNilMessage         = errors.New("message is nil")
	ErrInvalidCursor      = errors.New("invalid cursor")
	ErrInvalidLimit       = errors.New("invalid limit")
	ErrInvalidMessageID   = errors.New("invalid message id")
)
