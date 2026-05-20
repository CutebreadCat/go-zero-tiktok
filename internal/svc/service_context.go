// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"
	repository "go_zero-tiktok/internal/dal/repository"
	"go_zero-tiktok/internal/domain/websocket"
	"go_zero-tiktok/internal/infra/ai"
	wscache "go_zero-tiktok/internal/infra/cache/ws"
	"go_zero-tiktok/internal/infra/storage/aliyun"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  *wscache.RedisCache
	Rdb    *redis.Redis
	Dal    *repository.Repositories
	Hub    *websocket.Hub
	AIChat *websocket.AIChat
	MQ     *MQComponents
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	// 初始化阿里云配置
	aliyun.GetAliConfig()
	aliyun.AliInit()

	c := wscache.NewRedisCache(dal.Rdb)
	dalRepo := repository.NewRepositories(dal.Db, dal.Rdb)
	aiAgent, err := ai.NewAgent(context.Background())
	if err != nil {
		logx.Must(err)
	}
	aiChat := websocket.NewAIChat(aiAgent, c, dalRepo.User)
	// 创建 Hub（先不注入 writer）
	hub := websocket.NewHub(c, c, c, dalRepo.Chat, dalRepo.Chat, aiChat)

	// 创建 AI Agent 和 AIChat

	// 初始化 MQ 并注入 writer
	mq := InitMQ(config.Kafka, hub, aiChat)

	return &ServiceContext{
		Config: config,
		DB:     dal.Db,
		Rdb:    dal.Rdb,
		Dal:    dalRepo,
		Cache:  c,
		Hub:    hub,
		AIChat: aiChat,
		MQ:     mq,
	}
}
