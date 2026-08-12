package kafka

import "context"

type ConsumerHandler interface {
	Consume(ctx context.Context, msg *Event) error
}

// ConsumerHandlerFunc 将普通函数适配为 ConsumerHandler。
type ConsumerHandlerFunc func(ctx context.Context, msg *Event) error

func (f ConsumerHandlerFunc) Consume(ctx context.Context, msg *Event) error { return f(ctx, msg) }

// Consumer 消费路由：把事件交给具体的 ConsumerHandler 处理
type Consumer struct {
	Handler ConsumerHandler
}

func (c *Consumer) Consume(ctx context.Context, e *Event) error {
	if c.Handler == nil {
		return nil
	}
	return c.Handler.Consume(ctx, e)
}
