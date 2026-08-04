package kafka

import (
	"context"
	"fmt"
	"strings"

	appLogger "go_zero-tiktok/Prometheus/logger"

	"github.com/segmentio/kafka-go"
)

type TopicConfig struct {
	Name              string
	Partitions        int
	ReplicationFactor int
}

type MyKafa struct{}

func (m *MyKafa) BatchEnsureTopics(ctx context.Context, brokers []string, topics []TopicConfig) error {
	appLogger.Info("开始初始化 Kafka Topics...")

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
				appLogger.Infof("Topic [%s] 已存在，跳过", t.Name)
				continue
			}
			return fmt.Errorf("failed to create topic [%s]: %w", t.Name, err)
		}
		appLogger.Infof("Topic [%s] 创建成功 (分区:%d, 副本:%d)", t.Name, t.Partitions, t.ReplicationFactor)
	}

	appLogger.Info("all Kafka topics initialized")
	return nil
}
