package kafka

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestRetryConsumer_OnFailureFallbackSucceeds handler 重试耗尽后 OnFailure 降级成功，
// Consume 应返回 nil（视为已处理，可提交 offset），OnFailure 仅调用一次。
func TestRetryConsumer_OnFailureFallbackSucceeds(t *testing.T) {
	handler := ConsumerHandlerFunc(func(ctx context.Context, msg *Event) error {
		return errors.New("always fail")
	})

	var onFailureCalls int32
	cfg := RetryConfig{
		MaxRetry: 2,
		Backoff:  func(int) time.Duration { return 0 },
		OnFailure: func(ctx context.Context, msg *Event, err error) error {
			atomic.AddInt32(&onFailureCalls, 1)
			return nil // 降级成功
		},
	}
	rc := NewRetryConsumer(handler, cfg)

	msg := &Event{Msg: &Message{Key: []byte("k1")}}
	if err := rc.Consume(context.Background(), msg); err != nil {
		t.Fatalf("Consume should return nil when fallback succeeds, got %v", err)
	}
	if got := atomic.LoadInt32(&onFailureCalls); got != 1 {
		t.Fatalf("OnFailure should be called exactly once, got %d", got)
	}
}

// TestRetryConsumer_OnFailureFallbackFails handler 重试耗尽且 OnFailure 降级也失败，
// Consume 应返回 nil（走 DLQ/跳过），OnFailure 仅调用一次。
func TestRetryConsumer_OnFailureFallbackFails(t *testing.T) {
	handler := ConsumerHandlerFunc(func(ctx context.Context, msg *Event) error {
		return errors.New("always fail")
	})

	var onFailureCalls int32
	cfg := RetryConfig{
		MaxRetry: 1,
		Backoff:  func(int) time.Duration { return 0 },
		OnFailure: func(ctx context.Context, msg *Event, err error) error {
			atomic.AddInt32(&onFailureCalls, 1)
			return errors.New("fallback also failed")
		},
	}
	rc := NewRetryConsumer(handler, cfg)

	msg := &Event{Msg: &Message{Key: []byte("k1")}}
	if err := rc.Consume(context.Background(), msg); err != nil {
		t.Fatalf("Consume should return nil to avoid blocking consumption, got %v", err)
	}
	if got := atomic.LoadInt32(&onFailureCalls); got != 1 {
		t.Fatalf("OnFailure should be called exactly once, got %d", got)
	}
}

// TestRetryConsumer_SuccessSkipsOnFailure handler 重试后最终成功时不应触发 OnFailure。
func TestRetryConsumer_SuccessSkipsOnFailure(t *testing.T) {
	var attempts int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, msg *Event) error {
		if atomic.AddInt32(&attempts, 1) == 1 {
			return errors.New("fail once")
		}
		return nil
	})

	var onFailureCalls int32
	cfg := RetryConfig{
		MaxRetry: 3,
		Backoff:  func(int) time.Duration { return 0 },
		OnFailure: func(ctx context.Context, msg *Event, err error) error {
			atomic.AddInt32(&onFailureCalls, 1)
			return nil
		},
	}
	rc := NewRetryConsumer(handler, cfg)

	msg := &Event{Msg: &Message{Key: []byte("k1")}}
	if err := rc.Consume(context.Background(), msg); err != nil {
		t.Fatalf("Consume should return nil on success, got %v", err)
	}
	if got := atomic.LoadInt32(&onFailureCalls); got != 0 {
		t.Fatalf("OnFailure should NOT be called on success, got %d", got)
	}
}

// TestRetryConsumer_OnFailureOnExhaustion MaxRetry=1 时两次失败耗尽后触发 OnFailure 兜底。
func TestRetryConsumer_OnFailureOnExhaustion(t *testing.T) {
	handler := ConsumerHandlerFunc(func(ctx context.Context, msg *Event) error {
		return errors.New("always fail")
	})

	var onFailureCalls int32
	cfg := RetryConfig{
		MaxRetry: 1,
		Backoff:  func(int) time.Duration { return 0 },
		OnFailure: func(ctx context.Context, msg *Event, err error) error {
			atomic.AddInt32(&onFailureCalls, 1)
			return nil
		},
	}
	rc := NewRetryConsumer(handler, cfg)

	msg := &Event{Msg: &Message{Key: []byte("k1")}}
	if err := rc.Consume(context.Background(), msg); err != nil {
		t.Fatalf("Consume should return nil when fallback succeeds, got %v", err)
	}
	if got := atomic.LoadInt32(&onFailureCalls); got != 1 {
		t.Fatalf("OnFailure should be called exactly once, got %d", got)
	}
}
