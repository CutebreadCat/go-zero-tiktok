package kafka

import (
	"context"
	"encoding/json"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"

	"github.com/segmentio/kafka-go"
)

// producerErrQueueSize 异步写失败错误通道缓冲大小。
// kafka-go 通过 Completion 回调上报异步写结果，本封装把失败错误收集进该通道，
// 避免回调阻塞 writer 内部 goroutine。
const producerErrQueueSize = 64

type Producer struct {
	writer *kafka.Writer
	errCh  chan error
}

// ProducerConfig Producer 配置，业务可按场景选择同步/异步、Ack 级别
type ProducerConfig struct {
	Brokers      []string
	Topic        string
	Async        bool
	RequiredAcks int
	BatchSize    int
	BatchBytes   int
	BatchTimeout time.Duration
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	// Balancer 分区路由策略，决定消息进入哪个分区。
	// 业务需要"按 key 保证有序"（如点赞按 target_id 分区）时必须使用按 key 哈希的 Balancer：
	//   - &kafka.Hash{}：基于 FNV-1a 哈希，同 key 必进同一分区；key 为 nil 时退化为轮询。
	//   - &kafka.Murmur2Balancer{}：与 Java 客户端默认分区器一致。
	// 注意：&kafka.LeastBytes{}（旧默认）只按分区累计字节数做负载均衡、不按 key 哈希，
	// 同 key 消息不保证进同一分区，"按 target_id 保证有序"会失效。
	// 为空时默认使用 &kafka.Hash{}。
	Balancer kafka.Balancer
}

// NewProducerWithConfig 根据配置创建 Producer
func NewProducerWithConfig(cfg ProducerConfig) *Producer {
	if cfg.RequiredAcks == 0 {
		cfg.RequiredAcks = int(kafka.RequireOne)
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = producerBatchSize
	}
	if cfg.BatchBytes == 0 {
		cfg.BatchBytes = producerBatchBytes
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = producerBatchTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = producerWriteTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = producerReadTimeout
	}

	balancer := cfg.Balancer
	if balancer == nil {
		// 默认按 key 哈希分区：同 key（如点赞的 target_id）必进同一分区，保证单 target 有序。
		balancer = &kafka.Hash{}
	}

	// 异步模式下 kafka-go 通过 Completion 回调上报写结果，
	// 封装成错误通道，供调用方后台 goroutine 消费感知失败。
	// 注意：v0.4.51 的 kafka.WriterConfig 无 Completion 字段，需直接构造 Writer struct。
	errCh := make(chan error, producerErrQueueSize)
	completion := func(messages []kafka.Message, err error) {
		if err != nil {
			select {
			case errCh <- err:
			default:
				appLogger.Errorf("异步写 Kafka 失败且错误通道已满, 丢弃错误: %v", err)
			}
		}
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     balancer,
			WriteTimeout: cfg.WriteTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
			Async:        cfg.Async,
			BatchSize:    cfg.BatchSize,
			BatchBytes:   int64(cfg.BatchBytes),
			BatchTimeout: cfg.BatchTimeout,
			Completion:   completion,
		},
		errCh: errCh,
	}
}

// NewProducer 创建 Kafka Producer（异步模式，批量发送）
// 保留旧接口兼容，新业务优先使用 NewProducerWithConfig
func NewProducer(brokers []string, topic string) *Producer {
	return NewProducerWithConfig(ProducerConfig{
		Brokers:      brokers,
		Topic:        topic,
		Async:        true,
		RequiredAcks: int(kafka.RequireOne),
	})
}

func (k *Producer) SendMessage(ctx context.Context, m *Event) error {
	payload, err := k.MarshalMessage(m)
	if err != nil {
		return xerr.Wrap(err, "Producer.SendMessage.Marshal")
	}

	appLogger.Infof("发送消息到 Kafka, topic=%s, key=%s", m.Msg.Topic, string(m.Msg.Key))
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   m.Msg.Key,
		Value: payload,
	})
	if err != nil {
		return xerr.Wrap(err, "Producer.SendMessage.WriteMessages")
	}
	appLogger.Info("消息成功写入 Kafka")
	return nil
}

// Errors 返回异步写失败通道。
// 仅当 Async=true 时写失败才会进入该通道；调用方必须后台 goroutine 消费，
// 否则失败无法感知（或通道满时错误被丢弃并记日志）。Close 后通道被关闭，range 自动退出。
func (k *Producer) Errors() <-chan error {
	return k.errCh
}

// Close 关闭 Producer
func (k *Producer) Close() error {
	err := k.writer.Close()
	close(k.errCh)
	return err
}

// MarshalMessage 将整个 event 结构体序列化为 []byte（作为 Kafka 的 Value）
func (k *Producer) MarshalMessage(m *Event) ([]byte, error) {
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, xerr.Wrap(err, "Producer.MarshalMessage")
	}
	return payload, nil
}
