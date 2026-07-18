package mykafka

import (
	"context"
	"fmt"
	appLogger "go_zero-tiktok/Prometheus/logger"
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
	appLogger.Info("寮€濮嬪垵濮嬪寲 Kafka Topics...")

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
				appLogger.Infof("Topic [%s] 宸插瓨鍦紝璺宠繃", t.Name)
				continue
			}
			return fmt.Errorf("failed to create topic [%s]: %w", t.Name, err)
		}
		appLogger.Infof("Topic [%s] 鍒涘缓鎴愬姛 (鍒嗗尯:%d, 鍓湰:%d)", t.Name, t.Partitions, t.ReplicationFactor)
	}

	appLogger.Info("all Kafka topics initialized")
	return nil
}
