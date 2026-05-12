package mqcontract

import "context"

type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
}

type Event struct {
	Type string
	Msg  *Message
	Data any
}

type ConsumerHandler interface {
	Consume(ctx context.Context, msg *Event) error
}
