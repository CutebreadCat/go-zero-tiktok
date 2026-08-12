package kafka

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
)

// RetryConfig 消费重试配置
type RetryConfig struct {
	MaxRetry    int
	Backoff     func(retry int) time.Duration // 可选，为空时使用指数退避
	DLQProducer *Producer                     // 可选，死信队列生产者
	DLQTopic    string                        // 可选，死信 topic
}

// RetryConsumer 对 ConsumerHandler 包装有限次重试，超过后写入死信队列并返回 nil
// 返回 nil 是为了让上层正常提交 offset，避免单条 poison message 阻塞整个消费进度
type RetryConsumer struct {
	handler ConsumerHandler
	cfg     RetryConfig
}

func NewRetryConsumer(handler ConsumerHandler, cfg RetryConfig) *RetryConsumer {
	if cfg.MaxRetry <= 0 {
		cfg.MaxRetry = 3
	}
	if cfg.Backoff == nil {
		cfg.Backoff = defaultBackoff
	}
	return &RetryConsumer{
		handler: handler,
		cfg:     cfg,
	}
}

func (c *RetryConsumer) Consume(ctx context.Context, msg *Event) error {
	var lastErr error
	for i := 0; i <= c.cfg.MaxRetry; i++ {
		if err := c.safeConsume(ctx, msg); err == nil {
			return nil
		} else {
			lastErr = err
			appLogger.Warnf("consume failed, retry=%d, error=%v", i, err)
			if i < c.cfg.MaxRetry {
				time.Sleep(c.cfg.Backoff(i))
			}
		}
	}

	appLogger.Errorf("consume failed after %d retries, send to dlq or skip: %v", c.cfg.MaxRetry, lastErr)
	c.sendToDLQ(ctx, msg, lastErr)
	return nil
}

// safeConsume 把底层 handler 的 panic 转成 error（带堆栈），纳入重试流程，
// 避免 panic 逃逸导致调用方（如 worker goroutine）死亡。
func (c *RetryConsumer) safeConsume(ctx context.Context, msg *Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v\n%s", r, debug.Stack())
		}
	}()
	return c.handler.Consume(ctx, msg)
}

func (c *RetryConsumer) sendToDLQ(ctx context.Context, msg *Event, err error) {
	if c.cfg.DLQProducer == nil || c.cfg.DLQTopic == "" {
		return
	}

	dlqMsg := &Event{
		Type: "DeadLetter",
		Data: map[string]interface{}{
			"original": msg,
			"error":    err.Error(),
			"time":     time.Now().UnixMilli(),
		},
		Msg: &Message{Topic: c.cfg.DLQTopic},
	}

	if sendErr := c.cfg.DLQProducer.SendMessage(ctx, dlqMsg); sendErr != nil {
		logx.Errorf("send to dlq failed: %v", sendErr)
	}
}

func defaultBackoff(retry int) time.Duration {
	// 1s, 2s, 4s...
	d := time.Duration(1<<retry) * time.Second
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}
