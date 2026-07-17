package mqcontract

import "context"

type ConsumerHandler interface {
	Consume(ctx context.Context, msg *Event) error
}

type Router struct {
	unreadHandler  ConsumerHandler
	messageHandler ConsumerHandler
	roomHandler    ConsumerHandler
	aiChatHandler  ConsumerHandler
}

func NewRouter(unreadHandler, messageHandler, roomHandler, aiChatHandler ConsumerHandler) *Router {
	return &Router{
		unreadHandler:  unreadHandler,
		messageHandler: messageHandler,
		roomHandler:    roomHandler,
		aiChatHandler:  aiChatHandler,
	}
}

func (r *Router) Route(e *Event) ConsumerHandler {
	switch e.Type {
	case "get_unread":
		return r.unreadHandler
	case "message":
		return r.messageHandler
	case "room":
		return r.roomHandler
	case "ai_chat":
		return r.aiChatHandler
	default:
		return nil
	}
}

type Consumer struct {
	Router *Router
}

func (c *Consumer) Consume(ctx context.Context, e *Event) error {
	h := c.Router.Route(e)
	if h == nil {
		return nil
	}
	return h.Consume(ctx, e)
}
