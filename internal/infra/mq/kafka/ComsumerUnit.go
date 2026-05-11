package mykafka

import (
	"context"

	kafkaGo "github.com/segmentio/kafka-go"

	"go_zero-tiktok/internal/domain/mq"
)

type Reader struct {
	r *kafkaGo.Reader
}

func NewReader(r *kafkaGo.Reader) *Reader {
	return &Reader{r: r}
}

func (k *Reader) Fetch(ctx context.Context) (*mq.Message, error) {
	m, err := k.r.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	return &mq.Message{
		Topic:     m.Topic,
		Partition: m.Partition,
		Offset:    m.Offset,
		Key:       m.Key,
		Value:     m.Value,
	}, nil
}

func (k *Reader) Commit(ctx context.Context, msg *mq.Message) error {
	return k.r.CommitMessages(ctx, kafkaGo.Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
	})
}
