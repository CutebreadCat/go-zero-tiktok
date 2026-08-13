package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string `json:"DataSource"`

	// AppRedis 应用 Redis，用于 like_count 缓存与脏标记。
	AppRedis redis.RedisConf `json:"AppRedis"`
	// Kafka Kafka 配置，用于异步点赞事件。
	Kafka KafkaConfig `json:"Kafka"`
	// LikeSync Redis 兜底同步器配置；Kafka 消费异常时定期对齐 Redis 与 MySQL。
	LikeSync LikeSyncConfig `json:"LikeSync"`
}

// KafkaConfig Kafka 配置。
type KafkaConfig struct {
	Brokers []string `json:"Brokers"`
	Topic   string   `json:"Topic"`
	GroupID string   `json:"GroupID"`
	// Enable 是否启用 Kafka 异步点赞链路；false 时回退到同步更新。
	Enable bool `json:"Enable"`
	// WorkerCount 消费端 WorkerPool 并发数；0 时使用 pkg/kafka 默认值。
	WorkerCount int `json:"WorkerCount"`
	// QueueSize 消费端内部任务队列大小；0 时使用 pkg/kafka 默认值。
	QueueSize int `json:"QueueSize"`
}

// LikeSyncConfig Redis dirty-set 兜底同步器配置。
type LikeSyncConfig struct {
	// Enable 是否启用兜底同步器。
	Enable bool `json:"Enable"`
	// Interval 同步周期；如 "300s"。
	Interval time.Duration `json:"Interval"`
	// BatchSize 每次从 Redis 取出的脏视频数。
	BatchSize int `json:"BatchSize"`
}
