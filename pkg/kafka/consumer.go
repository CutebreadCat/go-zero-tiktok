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

// NewGroupReader 创建消费者组模式的 Reader，业务侧优先使用这个。
// CommitInterval 固定为 0（同步提交），确保 offset 提交节奏完全由 commitCoordinator 控制。
func NewGroupReader(brokers []string, topic, groupID string) *KafkaReader {
	r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       readerMinBytes,
		MaxBytes:       readerMaxBytes,
		CommitInterval: 0,
	})
	return NewReader(r)
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
	if msg == nil {
		return nil
	}

	appLogger.Infof("提交 Kafka offset: topic=%s, partition=%d, offset=%d", msg.Topic, msg.Partition, msg.Offset)
	if err := k.r.CommitMessages(ctx, kafkaGo.Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}); err != nil {
		return xerr.Wrap(err, "KafkaReader.Commit.CommitMessages")
	}
	return nil
}

// Close 关闭底层 Reader
func (k *KafkaReader) Close() error {
	return k.r.Close()
}

func (k *KafkaReader) UnmarshalMessage(m *Message) (*Event, error) {
	var event Event
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		return nil, xerr.Wrap(err, "KafkaReader.UnmarshalMessage")
	}
	return &event, nil
}
