package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string `json:"DataSource"`

	// AppRedis 应用 Redis，用于 Feed 候选池（feed:global ZSet）与热度排序（hot:videos ZSet）。
	AppRedis redis.RedisConf `json:"AppRedis"`

	// Hot 热度分计算与清理配置。
	Hot HotConfig `json:"Hot"`

	// QoS 播放质量聚合配置。
	QoS QoSConfig `json:"QoS"`

	// Kafka 配置，用于消费热度分重算事件与访问事件。
	Kafka KafkaConfig `json:"Kafka"`

	// Recommend 推荐策略配置。
	Recommend RecommendConfig `json:"Recommend"`
}

// RecommendConfig 推荐策略配置。
type RecommendConfig struct {
	// FetchFactor 召回倍数，每路召回 limit * FetchFactor 条候选。
	FetchFactor int `json:"FetchFactor"`
	// MaxAuthorRepeat 单作者每页最大重复次数。
	MaxAuthorRepeat int `json:"MaxAuthorRepeat"`
	// SeenTTL 曝光记录保留时长，0 表示使用默认值 7 天。
	SeenTTL time.Duration `json:"SeenTTL"`
	// SeenMaxSize 单个用户曝光记录最大容量，0 表示使用默认值 5000。
	SeenMaxSize int `json:"SeenMaxSize"`
	// HotWeight 热度分权重。
	HotWeight float64 `json:"HotWeight"`
	// RecencyWeight 时效性权重。
	RecencyWeight float64 `json:"RecencyWeight"`
	// FollowWeight 关注加权。
	FollowWeight float64 `json:"FollowWeight"`
	// QoSWeight QoS 权重。
	QoSWeight float64 `json:"QoSWeight"`
}

type QoSConfig struct {
	// AggregateInterval 聚合任务执行间隔，0 表示不启动。
	AggregateInterval time.Duration `json:"AggregateInterval"`
	// AggregateBatchSize 每次聚合读取的最大上报条数。
	AggregateBatchSize int32 `json:"AggregateBatchSize"`
	// LastIDRedisKey 记录上次处理到的上报 id 的 Redis key。
	LastIDRedisKey string `json:"LastIDRedisKey"`
}

// KafkaConfig Kafka 配置。
type KafkaConfig struct {
	Brokers []string `json:"Brokers"`
	// HotScoreTopic 热度分重算 topic（由 Gateway 发送）。
	HotScoreTopic string `json:"HotScoreTopic"`
	// VisitTopic 视频访问事件 topic（由本服务发送并消费）。
	VisitTopic string `json:"VisitTopic"`
	// GroupID 消费组 ID。
	GroupID string `json:"GroupID"`
	// Enable 是否启用 Kafka 事件链路。
	Enable bool `json:"Enable"`
	// WorkerCount 消费端 WorkerPool 并发数；0 时使用 pkg/kafka 默认值。
	WorkerCount int `json:"WorkerCount"`
	// QueueSize 消费端内部任务队列大小；0 时使用 pkg/kafka 默认值。
	QueueSize int `json:"QueueSize"`
}

type HotConfig struct {
	// BaseScore 热度分基础分，让新发布视频即使无互动也能进榜。
	BaseScore float64 `json:"BaseScore"`
	// LikeWeight 点赞权重。
	LikeWeight float64 `json:"LikeWeight"`
	// CommentWeight 评论权重。
	CommentWeight float64 `json:"CommentWeight"`
	// FavoriteWeight 收藏权重。
	FavoriteWeight float64 `json:"FavoriteWeight"`
	// Gravity HN 公式时间衰减指数。
	Gravity float64 `json:"Gravity"`
	// CleanupInterval 过期清理任务执行间隔。
	CleanupInterval time.Duration `json:"CleanupInterval"`
	// Window 滑动窗口时长，超过该时长未活跃的视频会被清理。
	Window time.Duration `json:"Window"`
	// KeepTopN 热度榜保留的最大视频数，超出部分会被裁剪。
	KeepTopN int `json:"KeepTopN"`
}
