package mykafka

import (
	"context"
	"log"
	"time"

	"encoding/json"

	mqcontract "go_zero-tiktok/internal/shared/mq"

	"github.com/segmentio/kafka-go"
)

type KafakaProducer struct {
	writer *kafka.Writer
}

// NewProducer 创建 Kafka Producer（独立构造函数）
func NewProducer(brokers []string, topic string) *KafakaProducer {
	return &KafakaProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      brokers,
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
			RequiredAcks: int(kafka.RequireOne),
		}),
	}
}

func (k *KafakaProducer) SendMessage(ctx context.Context, m *mqcontract.Event) error {
	var payload []byte
	var err error
	if payload, err = k.MashalMessage(m); err != nil {
		log.Printf("Failed to marshal message:%v", err)
		return err
	}
	for i := 0; i < 3; i++ {
		err := k.writer.WriteMessages(ctx, kafka.Message{
			Key:   m.Msg.Key,
			Value: payload,
			Topic: m.Msg.Topic,
		})
		if err != nil {
			if err == kafka.LeaderNotAvailable {
				log.Printf("Leader not available for topic %s, retrying... (%d/3)", m.Msg.Topic, i+1)
				time.Sleep(2 * time.Second)
				continue
			} else {
				log.Printf("Failed to send kafka message to topic %s: %v", m.Msg.Topic, err)
				return err
			}

		} else {
			break
		}
	}
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
