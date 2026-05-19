package mykafka

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

type MyKafa struct {
	producer *KafakaProducer

	reader *KafkaReader
}

func (m *MyKafa) BatchEnsureTopics(ctx context.Context, brokers []string, topics []TopicConfig) error {
	fmt.Println("开始初始化 Kafka Topics...")

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
				fmt.Printf("Topic [%s] 已存在，跳过\n", t.Name)
				continue
			}
			return fmt.Errorf("failed to create topic [%s]: %w", t.Name, err)
		}
		fmt.Printf("Topic [%s] 创建成功 (分区:%d, 副本:%d)\n", t.Name, t.Partitions, t.ReplicationFactor)
	}

	fmt.Println("所有 Topic 初始化完成")
	return nil
}
