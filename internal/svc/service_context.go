// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/domain/websocket"
	"go_zero-tiktok/internal/infra/ai"
	wscache "go_zero-tiktok/internal/infra/cache/ws"
	"go_zero-tiktok/internal/infra/storage/aliyun"
	"go_zero-tiktok/internal/middleware"
	"go_zero-tiktok/internal/middleware/government/breaker"
	"go_zero-tiktok/internal/middleware/government/limiter"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	DB        *gorm.DB
	Cache     *wscache.RedisCache
	Rdb       *redis.Redis
	Dal       *Repositories
	Hub       *websocket.Hub
	AIChat    *websocket.AIChat
	MQ        *MQComponents
	RateLimit rest.Middleware
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	// 初始化阿里云配置
	aliyun.GetAliConfig()
	aliyun.AliInit()

	c := wscache.NewRedisCache(dal.Rdb)
	dalRepo := NewRepositories(dal.Db, dal.Rdb)

	aiLimiter := limiter.New(dal.Rdb, ai.DefaultLimitSeconds, ai.DefaultLimitMaxRequests, ai.DefaultLimitKeyPrefix)
	wsLimiter := limiter.New(dal.Rdb, websocket.DefaultLimitSeconds, websocket.DefaultLimitMaxRequests, websocket.DefaultLimitKeyPrefix)
	aiBreaker := breaker.New()

	aiAgent, err := ai.NewAgent(context.Background(), aiLimiter, aiBreaker)
	if err != nil {
		logx.Must(err)
	}
	aiChat := websocket.NewAIChat(aiAgent, c)
	// 创建 Hub（先不注入 writer）
	hub := websocket.NewHub(c, c, c, dalRepo.Chat, dalRepo.Chat, aiChat, wsLimiter)

	// 创建 AI Agent 和 AIChat

	// 初始化 MQ 并注入 writer
	mq := InitMQ(config.Kafka, hub, aiChat)

	return &ServiceContext{
		Config:    config,
		DB:        dal.Db,
		Rdb:       dal.Rdb,
		Dal:       dalRepo,
		Cache:     c,
		Hub:       hub,
		AIChat:    aiChat,
		MQ:        mq,
		RateLimit: middleware.NewRateLimitMiddleware(limiter.New(dal.Rdb, middleware.DefaultRateLimitSeconds, middleware.DefaultRateLimitMaxRequests, middleware.DefaultRateLimitKeyPrefix)).Handle,
	}
}
