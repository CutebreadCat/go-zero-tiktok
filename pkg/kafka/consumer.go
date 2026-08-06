package kafka

import (
	"context"
	"encoding/json"

	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"

	kafkaGo "github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	r *kafkaGo.Reader
}

func NewReader(r *kafkaGo.Reader) *KafkaReader {
	return &KafkaReader{r: r}
}

// NewKafkaReader 创建指定 topic+groupID 的 Reader
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

func (k *KafkaReader) Fetch(ctx context.Context) (*Event, error) {
	m, err := k.r.FetchMessage(ctx)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.Fetch.FetchMessage")
	}

	var event Event
	err = json.Unmarshal(m.Value, &event)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.Fetch.Unmarshal")
	}

	if event.Msg == nil {
		event.Msg = &Message{}
	}
	event.Msg.Topic = m.Topic
	event.Msg.Partition = m.Partition
	event.Msg.Offset = m.Offset
	event.Msg.Key = m.Key
	return &event, nil
}

func (k *KafkaReader) Commit(ctx context.Context, msg *Message) error {
	appLogger.Infof("消息已处理: topic=%s, partition=%d, offset=%d", msg.Topic, msg.Partition, msg.Offset)
	return nil
}

func (k *KafkaReader) UnmarshalMessage(m *Message) (*Event, error) {
	var event Event
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.UnmarshalMessage")
	}
	return &event, nil
}
