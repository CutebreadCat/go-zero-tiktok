package mykafka

import (
	"context"

	"go_zero-tiktok/app/chat/domain/websocket"
	mqcontract "go_zero-tiktok/internal/shared/mq"
)

// MessageWriterAdapter 适配器，把 KafakaProducer 适配成 websocket.MessageWriter
type MessageWriterAdapter struct {
	producer *KafakaProducer
	topic    string
}

func NewMessageWriterAdapter(producer *KafakaProducer, topic string) *MessageWriterAdapter {
	return &MessageWriterAdapter{
		producer: producer,
		topic:    topic,
	}
}

func (a *MessageWriterAdapter) SendMessage(ctx context.Context, event *mqcontract.Event) error {
	return a.producer.SendMessage(ctx, event)
}

// 确保实现了 websocket.MessageWriter 接口
var _ websocket.MessageWriter = (*MessageWriterAdapter)(nil)
