package svc

import (
	"go_zero-tiktok/app/chat/rpc/internal/config"
	chatdomain "go_zero-tiktok/app/chat/rpc/internal/domain"
	wscache "go_zero-tiktok/app/chat/rpc/internal/infra/cache/ws"
	mykafka "go_zero-tiktok/app/chat/rpc/internal/infra/mq/kafka"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	Dal         *Repositories
	ChatService *chatdomain.ChatService
	Redis       *redis.Redis
	WSCache     *wscache.RedisCache
	MQ          *MQComponents
}

type MQComponents struct {
	Producer *mykafka.KafakaProducer
	Consumer *mykafka.MultiTopicConsumerUnit
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	// 初始化 Redis
	rds := redis.MustNewRedis(c.AppRedis)
	wsCache := wscache.NewRedisCache(rds)

	dalRepo := NewRepositories(db)

	// 初始化 MQ
	mq := initMQ(c.Kafka)

	return &ServiceContext{
		Config:      c,
		DB:          db,
		Dal:         dalRepo,
		ChatService: chatdomain.NewChatService(dalRepo.Chat),
		Redis:       rds,
		WSCache:     wsCache,
		MQ:          mq,
	}
}

func initMQ(cfg config.KafkaConfig) *MQComponents {
	logx.Infof("initializing MQ components, brokers=%v, topic=%s", cfg.Brokers, cfg.Topic)

	producer := mykafka.NewProducer(cfg.Brokers, cfg.Topic)

	return &MQComponents{
		Producer: producer,
	}
}
