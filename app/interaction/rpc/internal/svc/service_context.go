package svc

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/internal/cache"
	"go_zero-tiktok/app/interaction/rpc/internal/config"
	commentdomain "go_zero-tiktok/app/interaction/rpc/internal/domain"
	"go_zero-tiktok/app/interaction/rpc/internal/domain/interaction"
	"go_zero-tiktok/app/interaction/rpc/internal/worker"
	appkafka "go_zero-tiktok/pkg/kafka"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext 聚合 interaction RPC 的所有基础设施与领域服务。
type ServiceContext struct {
	Config             config.Config
	DB                 *gorm.DB
	Rdb                *redis.Redis
	Dal                *Repositories
	CommentService     *commentdomain.CommentService
	InteractionService *interaction.InteractionService

	likeCountSyncer   *worker.LikeCountSyncer
	consumerUnit      *appkafka.MultiTopicConsumerUnit
	likeEventProducer interaction.LikeEventProducer
}

// NewServiceContext 初始化 interaction RPC 服务上下文。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	rdb := redis.MustNewRedis(c.AppRedis)

	dalRepo := NewRepositories(db)

	likeCache := cache.NewLikeCountCache(rdb)

	// Kafka 生产者：未启用或未配置时传 nil，InteractionService 内部回退到同步更新。
	var producer interaction.LikeEventProducer
	if c.Kafka.Enable && len(c.Kafka.Brokers) > 0 && c.Kafka.Brokers[0] != "" {
		kafkaProducer := NewKafkaLikeEventProducer(c.Kafka.Brokers, c.Kafka.Topic)
		kafkaProducer.WatchErrors()
		producer = kafkaProducer
	}

	interactionService := interaction.NewInteractionService(
		dalRepo.VideoInteraction,
		dalRepo.VideoStat,
		likeCache,
		producer,
	)

	ctx := &ServiceContext{
		Config:             c,
		DB:                 db,
		Rdb:                rdb,
		Dal:                dalRepo,
		CommentService:     commentdomain.NewCommentService(dalRepo.Comment),
		InteractionService: interactionService,
		likeEventProducer:  producer,
	}

	// Kafka 消费者与后台 syncer 仅在启用时启动。
	if c.Kafka.Enable && len(c.Kafka.Brokers) > 0 && c.Kafka.Brokers[0] != "" {
		ctx.startKafkaConsumer(c)
	}

	// Redis dirty-set 兜底同步器：默认关闭，仅在 Kafka 消费异常/未启用时打开，定期对齐 Redis 与 MySQL。
	if c.LikeSync.Enable {
		opts := []worker.SyncerOption{}
		if c.LikeSync.Interval > 0 {
			opts = append(opts, worker.WithSyncInterval(c.LikeSync.Interval))
		}
		if c.LikeSync.BatchSize > 0 {
			opts = append(opts, worker.WithBatchSize(c.LikeSync.BatchSize))
		}
		ctx.likeCountSyncer = worker.NewLikeCountSyncer(likeCache, dalRepo.VideoInteraction, dalRepo.VideoStat, opts...)
		ctx.likeCountSyncer.Start(context.Background())
	}

	return ctx
}

// startKafkaConsumer 启动 Kafka 消费者：消费点赞/收藏事件并持久化到 MySQL。
func (ctx *ServiceContext) startKafkaConsumer(c config.Config) {
	handler := interaction.NewLikeEventHandler(ctx.Dal.VideoInteraction)

	workerCount := c.Kafka.WorkerCount
	if workerCount <= 0 {
		workerCount = appkafka.DefaultConsumerWorkerCount
	}
	queueSize := c.Kafka.QueueSize
	if queueSize <= 0 {
		queueSize = appkafka.DefaultConsumerQueueSize
	}

	configs := []appkafka.ConsumerTopicConfig{{Topic: c.Kafka.Topic}}
	ctx.consumerUnit = appkafka.NewMultiTopicConsumerUnitFromConfigs(
		configs,
		c.Kafka.Brokers,
		c.Kafka.GroupID,
		handler,
		workerCount,
		queueSize,
		appkafka.RetryConfig{
			MaxRetry: 3,
			OnFailure: func(c context.Context, msg *appkafka.Event, err error) error {
				// 消费失败兜底：记录日志并告警。最终一致性由 LikeSync（Redis dirty-set 同步器）保证。
				logx.Errorf("kafka like event consume failed after retries: %v", err)
				return nil
			},
		},
	)
	ctx.consumerUnit.Start(context.Background())
}

// Close 优雅关闭后台资源。
func (ctx *ServiceContext) Close() {
	if ctx.likeCountSyncer != nil {
		ctx.likeCountSyncer.Stop()
	}
	if ctx.consumerUnit != nil {
		ctx.consumerUnit.Stop()
	}
	if ctx.likeEventProducer != nil {
		_ = ctx.likeEventProducer.Close()
	}
}
