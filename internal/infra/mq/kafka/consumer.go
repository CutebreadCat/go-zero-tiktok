package mykafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafakaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaReader(ctx context.Context, topic, groupID string) *KafakaConsumer {
	var k = &KafakaConsumer{}
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         groupID,
		Topic:           topic,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		ReadLagInterval: -1,   // 禁止自动更新滞后信息
	})
	return k
}
func (k *KafakaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		log.Printf("Failed to read kafka message: %v", err)
		return kafka.Message{}, err
	}
	return msg, nil
}
