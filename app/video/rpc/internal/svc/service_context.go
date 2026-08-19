package svc

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/config"
	videodomain "go_zero-tiktok/app/video/rpc/internal/domain"
	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/app/video/rpc/internal/domain/worker"
	"go_zero-tiktok/pkg/event"
	appkafka "go_zero-tiktok/pkg/kafka"
	"go_zero-tiktok/pkg/storage/aliyun"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext 聚合 video RPC 的所有基础设施与领域服务。
type ServiceContext struct {
	Config            config.Config
	DB                *gorm.DB
	Redis             *redis.Redis
	Dal               *Repositories
	VideoService      *videodomain.VideoService
	Storage           *StorageAdapter
	HotScoreCleaner   *worker.HotScoreCleaner
	consumerUnit      *appkafka.MultiTopicConsumerUnit
	visitProducer     videodomain.VisitEventProducer
}

// NewServiceContext 初始化 video RPC 服务上下文。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	// 初始化应用 Redis（Feed 候选池 feed:global）
	rdb := redis.MustNewRedis(c.AppRedis)

	// 初始化阿里云配置
	aliyun.LoadConfig()
	aliyun.InitClient()

	dalRepo, err := NewRepositories(db, rdb)
	logx.Must(err)
	storageAdapter := &StorageAdapter{}

	// Kafka 访问事件生产者：未启用或未配置时留空，RecordVisit 回退到同步处理。
	var visitProducer videodomain.VisitEventProducer
	if c.Kafka.Enable && len(c.Kafka.Brokers) > 0 && c.Kafka.Brokers[0] != "" {
		producer := NewKafkaVideoVisitProducer(c.Kafka.Brokers, c.Kafka.VisitTopic)
		producer.WatchErrors()
		visitProducer = producer
	}

	hotCalculator := newHotScoreCalculator(c.Hot)
	videoService := videodomain.NewVideoService(dalRepo.Video, dalRepo.VideoStat, storageAdapter, dalRepo.Feed, hotCalculator, visitProducer)

	ctx := &ServiceContext{
		Config:        c,
		DB:            db,
		Redis:         rdb,
		Dal:           dalRepo,
		VideoService:  videoService,
		Storage:       storageAdapter,
		visitProducer: visitProducer,
	}

	// 启动热度分清理 worker（定时删除过期成员并裁剪规模）。
	if c.Hot.CleanupInterval > 0 && c.Hot.Window > 0 {
		cleanerOpts := []worker.CleanerOption{
			worker.WithCleanupInterval(c.Hot.CleanupInterval),
			worker.WithCleanupWindow(c.Hot.Window),
		}
		if c.Hot.KeepTopN > 0 {
			cleanerOpts = append(cleanerOpts, worker.WithKeepTopN(c.Hot.KeepTopN))
		}
		ctx.HotScoreCleaner = worker.NewHotScoreCleaner(dalRepo.Feed, cleanerOpts...)
		ctx.HotScoreCleaner.Start(context.Background())
	}

	// Kafka 热度事件消费者：消费热度分重算事件与访问事件。
	if c.Kafka.Enable && len(c.Kafka.Brokers) > 0 && c.Kafka.Brokers[0] != "" {
		ctx.startKafkaConsumer(c, videoService)
	}

	return ctx
}

// startKafkaConsumer 启动 Kafka 消费：热度分重算事件 + 访问事件。
func (ctx *ServiceContext) startKafkaConsumer(c config.Config, videoService *videodomain.VideoService) {
	handler := videodomain.NewHotScoreEventHandler(videoService)

	workerCount := c.Kafka.WorkerCount
	if workerCount <= 0 {
		workerCount = appkafka.DefaultConsumerWorkerCount
	}
	queueSize := c.Kafka.QueueSize
	if queueSize <= 0 {
		queueSize = appkafka.DefaultConsumerQueueSize
	}

	topics := make([]appkafka.ConsumerTopicConfig, 0, 2)
	if c.Kafka.HotScoreTopic != "" {
		topics = append(topics, appkafka.ConsumerTopicConfig{Topic: c.Kafka.HotScoreTopic})
	}
	if c.Kafka.VisitTopic != "" {
		topics = append(topics, appkafka.ConsumerTopicConfig{Topic: c.Kafka.VisitTopic})
	}
	if len(topics) == 0 {
		topics = []appkafka.ConsumerTopicConfig{
			{Topic: event.DefaultHotScoreRecalcTopic},
			{Topic: event.DefaultVideoVisitTopic},
		}
	}

	ctx.consumerUnit = appkafka.NewMultiTopicConsumerUnitFromConfigs(
		topics,
		c.Kafka.Brokers,
		c.Kafka.GroupID,
		handler,
		workerCount,
		queueSize,
		appkafka.RetryConfig{
			MaxRetry: 3,
			OnFailure: func(c context.Context, msg *appkafka.Event, err error) error {
				logx.Errorf("kafka hot-score event consume failed after retries: %v", err)
				return nil
			},
		},
	)
	ctx.consumerUnit.Start(context.Background())
}

const (
	defaultHotBaseScore      = 100
	defaultHotLikeWeight     = 5
	defaultHotCommentWeight  = 10
	defaultHotFavoriteWeight = 3
	defaultHotGravity        = 1.5
)

// newHotScoreCalculator 根据配置创建热度分计算器，未配置时使用默认值。
func newHotScoreCalculator(c config.HotConfig) *feedpkg.HotScoreCalculator {
	calc := &feedpkg.HotScoreCalculator{
		BaseScore:      defaultHotBaseScore,
		LikeWeight:     defaultHotLikeWeight,
		CommentWeight:  defaultHotCommentWeight,
		FavoriteWeight: defaultHotFavoriteWeight,
		Gravity:        defaultHotGravity,
	}
	if c.BaseScore != 0 {
		calc.BaseScore = c.BaseScore
	}
	if c.LikeWeight != 0 {
		calc.LikeWeight = c.LikeWeight
	}
	if c.CommentWeight != 0 {
		calc.CommentWeight = c.CommentWeight
	}
	if c.FavoriteWeight != 0 {
		calc.FavoriteWeight = c.FavoriteWeight
	}
	if c.Gravity != 0 {
		calc.Gravity = c.Gravity
	}
	return calc
}

// Close 优雅关闭后台资源。
func (ctx *ServiceContext) Close() {
	if ctx.HotScoreCleaner != nil {
		ctx.HotScoreCleaner.Stop()
	}
	if ctx.consumerUnit != nil {
		ctx.consumerUnit.Stop()
	}
	if ctx.visitProducer != nil {
		_ = ctx.visitProducer.Close()
	}
}
