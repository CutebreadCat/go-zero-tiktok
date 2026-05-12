package svc

import (
	"context"
	"log"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/domain/websocket"
	"go_zero-tiktok/internal/infra/cache"
	mykafka "go_zero-tiktok/internal/infra/mq/kafka"
	mqcontract "go_zero-tiktok/internal/shared/mq"
)

// MQComponents MQ 组件集合
type MQComponents struct {
	Producer *mykafka.KafakaProducer
	Consumer *mykafka.MultiTopicConsumerUnit
}

// InitMQ 初始化 MQ 组件
func InitMQ(cfg config.KafkaConfig, hub *websocket.Hub, c *cache.RedisCache) *MQComponents {
	// 1. 创建 Kafka Producer
	producer := mykafka.NewProducer(cfg.Brokers, cfg.Topic)

	// 2. 创建 MessageWriter 适配器
	writer := mykafka.NewMessageWriterAdapter(producer, cfg.Topic)

	// 3. 注入 writer 到 Hub
	hub.SetWriter(writer)

	// 4. 创建业务 Handler
	messageHandler := websocket.NewMessageHandler(hub.Messages())
	unreadHandler := websocket.NewUnreadHandler(hub.Messages())

	// 5. 创建 Router 和 Consumer
	router := mqcontract.NewRouter(unreadHandler, messageHandler, nil)
	consumer := &mqcontract.Consumer{Router: router}

	// 6. 创建消费单元
	consumerUnit := mykafka.NewMultiTopicConsumerUnitFromConfigs(
		[]mykafka.ConsumerTopicConfig{
			{Topic: cfg.Topic, PartitionCount: 3},
		},
		cfg.GroupID,
		consumer,
		10,   // Worker 数量
		1000, // 队列大小
	)

	// 7. 启动消费
	consumerUnit.Start(context.Background())
	log.Println("MQ Consumer started")

	return &MQComponents{
		Producer: producer,
		Consumer: consumerUnit,
	}
}
