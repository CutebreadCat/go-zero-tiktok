package mykafka

import (
	"context"
	mqcontract "go_zero-tiktok/internal/shared/mq"
	"log"

	kafkaGo "github.com/segmentio/kafka-go"

	"encoding/json"
)

type KafkaReader struct {
	r *kafkaGo.Reader
}

func NewReader(r *kafkaGo.Reader) *KafkaReader {
	return &KafkaReader{r: r}
}

func (m *MyKafa) NewKafkaReader(ctx context.Context, brokers []string, topic string, groupID string) *KafkaReader {
	r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return NewReader(r)
}

func (k *KafkaReader) Fetch(ctx context.Context) (*mqcontract.Event, error) {
	m, err := k.r.FetchMessage(ctx)
	if err != nil {
		log.Printf("❌ KafkaReader FetchMessage error: %v", err)
		return nil, err
	}

	log.Printf("📨 KafkaReader fetched message: topic=%s, partition=%d, offset=%d, key=%s, value_len=%d",
		m.Topic, m.Partition, m.Offset, string(m.Key), len(m.Value))

	var Event mqcontract.Event
	err = json.Unmarshal(m.Value, &Event)
	if err != nil {
		log.Printf("❌ Failed to unmarshal message: %v", err)
		return nil, err
	}

	Event.Msg.Topic = m.Topic
	Event.Msg.Partition = m.Partition
	Event.Msg.Offset = m.Offset
	Event.Msg.Key = m.Key
	return &Event, nil
}

func (k *KafkaReader) Commit(ctx context.Context, msg *mqcontract.Message) error {
	// 当使用 Partition 模式（没有 GroupID）时，不需要手动 commit
	// kafka-go 会自动跟踪 offset
	log.Printf("✅ Message processed: topic=%s, partition=%d, offset=%d", msg.Topic, msg.Partition, msg.Offset)
	return nil
}

func (k *KafkaReader) UnmarshalMessage(m *mqcontract.Message) (*mqcontract.Event, error) {
	var event mqcontract.Event
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		log.Printf("Failed to unmarshal kafka message: %v", err)
		return nil, err
	}
	return &event, nil
}
