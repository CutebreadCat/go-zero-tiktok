package svc

import (
	"context"
	"log"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/app/chat/domain/websocket"
	mykafka "go_zero-tiktok/internal/infra/mq/kafka"
	mqcontract "go_zero-tiktok/internal/shared/mq"
)

type MQComponents struct {
	Producer *mykafka.KafakaProducer
	Consumer *mykafka.MultiTopicConsumerUnit
}

func InitMQ(cfg config.KafkaConfig, hub *websocket.Hub, aiChat *websocket.AIChat) *MQComponents {
	log.Printf("initializing MQ components")
	log.Printf("Kafka producer config: brokers=%v, topic=%s", cfg.Brokers, cfg.Topic)
	log.Printf("Kafka consumer config: brokers=%v, topic=%s, groupID=%s", cfg.Brokers, cfg.Topic, cfg.GroupID)

	mqcontract.RegisterEventFactory(websocket.EventTypeMessage, func() any { return &websocket.MessageEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeUnread, func() any { return &websocket.UnreadEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeRoom, func() any { return &websocket.RoomEvent{} })
	mqcontract.RegisterEventFactory(websocket.EventTypeAIChat, func() any { return &websocket.AIChatEvent{} })

	producer := mykafka.NewProducer(cfg.Brokers, cfg.Topic)
	writer := mykafka.NewMessageWriterAdapter(producer, cfg.Topic)
	hub.SetWriter(writer)

	messageHandler := websocket.NewMessageHandler(hub.Messages())
	unreadHandler := websocket.NewUnreadHandler(hub.Messages())
	aiChatHandler := websocket.NewAIChatHandler(aiChat, hub.Rooms())

	router := mqcontract.NewRouter(unreadHandler, messageHandler, nil, aiChatHandler)
	consumer := &mqcontract.Consumer{Router: router}
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

	consumerUnit.Start(context.Background())
	log.Println("MQ consumer started")

	return &MQComponents{
		Producer: producer,
		Consumer: consumerUnit,
	}
}
