package mqcontract

import "context"

type Router struct {
	unreadHandler  ConsumerHandler
	messageHandler ConsumerHandler
	roomHandler    ConsumerHandler
}

func NewRouter(unreadHandler, messageHandler, roomHandler ConsumerHandler) *Router {
	return &Router{
		unreadHandler:  unreadHandler,
		messageHandler: messageHandler,
		roomHandler:    roomHandler,
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
