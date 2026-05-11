package mq

import "context"

type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
}
type Event struct {
	Type string
	msg  *Message
}

type ConsumerHandler interface {
	Consume(ctx context.Context, msg *Event) error
}

type Router struct {
	unreadHandler ConsumerHandler
	aiHandler     ConsumerHandler
	roomHandler   ConsumerHandler
}

func (r *Router) Route(e *Event) ConsumerHandler {
	switch e.Type {
	case "get_unread":
		return r.unreadHandler
	case "message":
		return r.aiHandler
	case "room":
		return r.roomHandler
	default:
		return nil
	}
}

func (r *Router) Consume(ctx context.Context, e *Event) error {
	return r.Route(e).Consume(ctx, e)
}
