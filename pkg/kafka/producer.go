package kafka

import (
	"context"
	"encoding/json"

	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/pkg/xerr"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

// NewProducer 创建 Kafka Producer（异步模式，批量发送）
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
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

func (k *Producer) SendMessage(ctx context.Context, m *Event) error {
	payload, err := k.MarshalMessage(m)
	if err != nil {
		return xerr.Wrap(err, "Producer.SendMessage.Marshal")
	}

	appLogger.Infof("发送消息到 Kafka, topic=%s, key=%s", m.Msg.Topic, string(m.Msg.Key))
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   m.Msg.Key,
		Value: payload,
	})
	if err != nil {
		return xerr.Wrap(err, "Producer.SendMessage.WriteMessages")
	}
	appLogger.Info("消息成功写入 Kafka")
	return nil
}

// MarshalMessage 将整个 event 结构体序列化为 []byte（作为 Kafka 的 Value）
func (k *Producer) MarshalMessage(m *Event) ([]byte, error) {
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, xerr.Wrap(err, "Producer.MarshalMessage")
	}
	return payload, nil
}
