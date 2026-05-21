package svc

import (
	"context"
	"log"
	"net"
	"time"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/domain/websocket"
	mykafka "go_zero-tiktok/internal/infra/mq/kafka"
	mqcontract "go_zero-tiktok/internal/shared/mq"
)

// waitForKafka 等待 Kafka 连接就绪，最多重试 30 次，每次间隔 2 秒
func waitForKafka(broker string, maxRetries int) error {
	log.Printf("⏳ 等待 Kafka 连接就绪: %s", broker)
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", broker, 2*time.Second)
		if err == nil {
			conn.Close()
			log.Printf("✅ Kafka 连接成功！")
			return nil
		}
		log.Printf("⏳ Kafka 未就绪，等待重试 (%d/%d)...", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return nil
}

// MQComponents MQ 组件集合
type MQComponents struct {
	Producer *mykafka.KafakaProducer
	Consumer *mykafka.MultiTopicConsumerUnit
}

// InitMQ 初始化 MQ 组件
func InitMQ(cfg config.KafkaConfig, hub *websocket.Hub, aiChat *websocket.AIChat) *MQComponents {
	log.Printf("🚀 初始化 MQ 组件...")
	log.Printf("Kafka Producer 配置：Brokers=%v, Topic=%s", cfg.Brokers, cfg.Topic)
	log.Printf("Kafka Consumer 配置：Brokers=%v, Topic=%s, GroupID=%s", cfg.Brokers, cfg.Topic, cfg.GroupID)

	// 0. 注册事件工厂，使 Kafka 反序列化时自动还原为具体类型
	mqcontract.RegisterEventFactory(websocket.EventTypeMessage, func() any { return &websocket.MessageEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeUnread, func() any { return &websocket.UnreadEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeRoom, func() any { return &websocket.RoomEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeAIChat, func() any { return &websocket.AIChatEvent{} })

	// 1. 创建 Kafka Producer
	producer := mykafka.NewProducer(cfg.Brokers, cfg.Topic)
	log.Printf("✅ Kafka Producer 创建成功")

	// 2. 创建 MessageWriter 适配器
	writer := mykafka.NewMessageWriterAdapter(producer, cfg.Topic)

	// 3. 注入 writer 到 Hub
	hub.SetWriter(writer)
	log.Printf("✅ Hub Writer 注入成功")

	// 4. 创建业务 Handler
	messageHandler := websocket.NewMessageHandler(hub.Messages())
	unreadHandler := websocket.NewUnreadHandler(hub.Messages())
	aiChatHandler := websocket.NewAIChatHandler(aiChat, hub.Rooms())
	log.Printf("✅ MessageHandler、UnreadHandler 和 AIChatHandler 创建成功")

	// 5. 创建 Router 和 Consumer
	router := mqcontract.NewRouter(unreadHandler, messageHandler, nil, aiChatHandler)
	consumer := &mqcontract.Consumer{Router: router}
	log.Printf("✅ Router 和 Consumer 创建成功")

	// 6. 创建消费单元
	log.Printf("🚀 创建 MultiTopicConsumerUnit...")
	consumerUnit := mykafka.NewMultiTopicConsumerUnitFromConfigs(
		[]mykafka.ConsumerTopicConfig{
			{Topic: cfg.Topic},
		},
		cfg.Brokers,
		cfg.GroupID,
		consumer,
		mykafka.DefaultConsumerWorkerCount,
		mykafka.DefaultConsumerQueueSize,
	)

	// 7. 启动消费
	log.Printf("🚀 启动 ConsumerUnit...")
	consumerUnit.Start(context.Background())
	log.Println("✅ MQ Consumer started")

	return &MQComponents{
		Producer: producer,
		Consumer: consumerUnit,
	}
}
