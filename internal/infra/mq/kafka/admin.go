package mykafka

import (
	"context"
	"fmt"
	"log"
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
	log.Println("🚀 开始初始化 Kafka Topics...")

	// 1. 建立连接 (只建立一次，复用连接)
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial broker: %w", err)
	}
	defer conn.Close()

	// 2. 循环创建
	for _, t := range topics {
		config := kafka.TopicConfig{
			Topic:             t.Name,
			NumPartitions:     t.Partitions,
			ReplicationFactor: t.ReplicationFactor,
		}

		err := conn.CreateTopics(config)
		if err != nil {
			// 忽略 "TopicAlreadyExists" 错误
			if strings.Contains(err.Error(), "TopicAlreadyExists") {
				log.Printf("ℹ️ Topic [%s] 已存在，跳过", t.Name)
				continue
			}
			// 其他错误（如权限不足、参数非法）则返回
			return fmt.Errorf("failed to create topic [%s]: %w", t.Name, err)
		}
		log.Printf("✅ Topic [%s] 创建成功 (分区:%d, 副本:%d)", t.Name, t.Partitions, t.ReplicationFactor)
	}

	log.Println("🎉 所有 Topic 初始化完成")
	return nil
}
