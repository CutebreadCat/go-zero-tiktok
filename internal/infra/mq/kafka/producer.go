package mykafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafakaProducer struct {
	writer *kafka.Writer
}

func (m *MyKafa) NewKafkaWriter(ctx context.Context, topic string) *KafakaProducer {
	var k = &KafakaProducer{}
	k.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		WriteTimeout: 10 * 1000 * 1000 * 1000, // 10秒
		ReadTimeout:  10 * time.Second,
		RequiredAcks: int(kafka.RequireOne), // 等待所有副本确认

	})
	return k
}

func (k *KafakaProducer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	message := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	for i := 0; i < 3; i++ {
		err := k.writer.WriteMessages(ctx, message)
		if err != nil {
			if err == kafka.LeaderNotAvailable {
				log.Printf("Leader not available for topic %s, retrying... (%d/3)", topic, i+1)
				time.Sleep(2 * time.Second)
				continue
			} else {
				log.Printf("Failed to send kafka message to topic %s: %v", topic, err)
				return err
			}

		} else {
			break
		}
	}
	return nil

}
