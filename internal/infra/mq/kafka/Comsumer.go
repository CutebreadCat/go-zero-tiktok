package mykafka

import (
	"context"
	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/svc/xerr"
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
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return NewReader(r)
}

func (k *KafkaReader) Fetch(ctx context.Context) (*mqcontract.Event, error) {
	m, err := k.r.FetchMessage(ctx)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.Fetch.FetchMessage")
	}

	log.Printf("KafkaReader 获取消息: topic=%s, partition=%d, offset=%d, key=%s, value_len=%d",
		m.Topic, m.Partition, m.Offset, string(m.Key), len(m.Value))

	var Event mqcontract.Event
	err = json.Unmarshal(m.Value, &Event)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.fetch.Unmarshal")
	}

	Event.Msg.Topic = m.Topic
	Event.Msg.Partition = m.Partition
	Event.Msg.Offset = m.Offset
	Event.Msg.Key = m.Key
	return &Event, nil
}

func (k *KafkaReader) Commit(ctx context.Context, msg *mqcontract.Message) error {
	log.Printf("消息已处理: topic=%s, partition=%d, offset=%d", msg.Topic, msg.Partition, msg.Offset)
	return nil
}

func (k *KafkaReader) UnmarshalMessage(m *mqcontract.Message) (*mqcontract.Event, error) {
	var event mqcontract.Event
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.UnmarshalMessage")
	}
	return &event, nil
}
