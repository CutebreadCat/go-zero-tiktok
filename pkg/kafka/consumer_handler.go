package kafka

import "context"

type ConsumerHandler interface {
	Consume(ctx context.Context, msg *Event) error
}

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
