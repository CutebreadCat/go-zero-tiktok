package mykafka

import (
	"context"
	"log"

	"encoding/json"

	mqcontract "go_zero-tiktok/internal/shared/mq"

	"github.com/segmentio/kafka-go"
)

type KafakaProducer struct {
	writer *kafka.Writer
}

// NewProducer 创建 Kafka Producer（异步模式，批量发送）
func NewProducer(brokers []string, topic string) *KafakaProducer {
	return &KafakaProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      brokers,
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: producerWriteTimeout,
			ReadTimeout:  producerReadTimeout,
			RequiredAcks: int(kafka.RequireOne),
			Async:        true, // 异步模式
			BatchSize:    producerBatchSize,
			BatchBytes:   producerBatchBytes,
			BatchTimeout: producerBatchTimeout,
		}),
	}
}

func (k *KafakaProducer) SendMessage(ctx context.Context, m *mqcontract.Event) error {
	var payload []byte
	var err error
	if payload, err = k.MashalMessage(m); err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	log.Printf("🚀 Writing message to Kafka, topic=%s, key=%s", m.Msg.Topic, string(m.Msg.Key))
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   m.Msg.Key,
		Value: payload,
	})
	if err != nil {
		log.Printf("❌ Failed to write kafka message: %v", err)
		return err
	}
	log.Printf("✅ Message written to Kafka successfully")
	return nil
}

func (k *KafakaProducer) MashalMessage(m *mqcontract.Event) ([]byte, error) {

	// 2. 将整个 event 结构体序列化为 []byte (作为 Kafka 的 Value)
	payload, err := json.Marshal(m)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, err
	}

	return payload, nil

}
