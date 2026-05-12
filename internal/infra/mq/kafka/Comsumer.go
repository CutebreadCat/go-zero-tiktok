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

func (m *MyKafa) NewKafkaReader(ctx context.Context, topic string, groupID string) *KafkaReader {
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
		log.Printf("Failed to fetch kafka message: %v", err)
		return nil, err
	}
	var Event mqcontract.Event
	err = json.Unmarshal(m.Value, &Event)
	if err != nil {
		log.Printf("Failed to unmarshal kafka message: %v", err)
		return nil, err
	}
	Event.Msg.Topic = m.Topic
	Event.Msg.Partition = m.Partition
	Event.Msg.Offset = m.Offset
	Event.Msg.Key = m.Key
	// 不在这里 Commit，交给 Partition 层的回调，处理成功后才提交 Offset
	return &Event, nil
}

func (k *KafkaReader) Commit(ctx context.Context, msg *mqcontract.Message) error {
	return k.r.CommitMessages(ctx, kafkaGo.Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
	})
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
