package mykafka

import (
	"context"
	"go_zero-tiktok/pkg/xerr"
	"log"

	"encoding/json"

	mqcontract "go_zero-tiktok/pkg/mq"

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
			Async:        true,
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
		return xerr.Wrap(err, "KafakaProducer.SendMessage.Marshal")
	}

	log.Printf("发送消息到 Kafka, topic=%s, key=%s", m.Msg.Topic, string(m.Msg.Key))
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   m.Msg.Key,
		Value: payload,
	})
	if err != nil {
		return xerr.Wrap(err, "KafakaProducer.SendMessage.WriteMessages")
	}
	log.Printf("消息成功写入 Kafka")
	return nil
}

func (k *KafakaProducer) MashalMessage(m *mqcontract.Event) ([]byte, error) {

	// 2. 将整个 event 结构体序列化为 []byte (作为 Kafka 的 Value)
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, xerr.Wrap(err, "KafakaProducer.MashalMessage")
	}

	return payload, nil

}
