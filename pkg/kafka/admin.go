package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

type TopicConfig struct {
	Name              string
	Partitions        int
	ReplicationFactor int
}

type MyKafa struct{}

func (m *MyKafa) BatchEnsureTopics(ctx context.Context, brokers []string, topics []TopicConfig) error {
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial broker: %w", err)
	}
	defer conn.Close()

	for _, t := range topics {
		config := kafka.TopicConfig{
			Topic:             t.Name,
			NumPartitions:     t.Partitions,
			ReplicationFactor: t.ReplicationFactor,
		}

		err := conn.CreateTopics(config)
		if err != nil {
			if strings.Contains(err.Error(), "TopicAlreadyExists") {
				continue
			}
			return fmt.Errorf("failed to create topic [%s]: %w", t.Name, err)
		}
	}

	return nil
}
