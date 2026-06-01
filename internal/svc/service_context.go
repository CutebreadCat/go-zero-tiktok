// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb/chat_pb"
	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	"go_zero-tiktok/app/user/rpc/userservice"
	videopb "go_zero-tiktok/app/video/rpc/video_pb/video_pb"
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
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Cache  *wscache.RedisCache
	Dal    *Repositories

	// RPC Clients
	UserRpc          userservice.UserService
	VideoRpc         videopb.VideoServiceClient
	InteractionRpc   interactionpb.InteractionServiceClient
	CommunicationRpc communicationpb.CommunicationServiceClient
	ChatRpc          chatpb.ChatServiceClient

	// WebSocket 相关
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
	dalRepo := NewRepositories(dal.Db)

	// 创建 RPC clients
	videoRpc := videopb.NewVideoServiceClient(zrpc.MustNewClient(config.VideoRpc).Conn())
	interactionRpc := interactionpb.NewInteractionServiceClient(zrpc.MustNewClient(config.InteractionRpc).Conn())
	communicationRpc := communicationpb.NewCommunicationServiceClient(zrpc.MustNewClient(config.CommunicationRpc).Conn())
	chatRpc := chatpb.NewChatServiceClient(zrpc.MustNewClient(config.ChatRpc).Conn())

	aiLimiter := limiter.New(dal.Rdb, ai.DefaultLimitSeconds, ai.DefaultLimitMaxRequests, ai.DefaultLimitKeyPrefix)
	wsLimiter := limiter.New(dal.Rdb, websocket.DefaultLimitSeconds, websocket.DefaultLimitMaxRequests, websocket.DefaultLimitKeyPrefix)
	aiBreaker := breaker.New()

	aiAgent, err := ai.NewAgent(context.Background(), aiLimiter, aiBreaker)
	if err != nil {
		logx.Must(err)
	}
	aiChat := websocket.NewAIChat(aiAgent, c)

	// 创建 Chat 仓储适配器
	chatRepoAdapter := NewChatRepoAdapter(chatRpc)

	// 创建 Hub
	hub := websocket.NewHub(c, c, c, chatRepoAdapter, chatRepoAdapter, aiChat, wsLimiter)

	// 初始化 MQ 并注入 writer
	mq := InitMQ(config.Kafka, hub, aiChat)

	return &ServiceContext{
		Config: config,
		Dal:    dalRepo,

		// RPC Clients
		UserRpc:          userservice.NewUserService(zrpc.MustNewClient(config.UserRpc)),
		VideoRpc:         videoRpc,
		InteractionRpc:   interactionRpc,
		CommunicationRpc: communicationRpc,
		ChatRpc:          chatRpc,

		// WebSocket 相关
		Cache:     c,
		Hub:       hub,
		AIChat:    aiChat,
		MQ:        mq,
		RateLimit: middleware.NewRateLimitMiddleware(limiter.New(dal.Rdb, middleware.DefaultRateLimitSeconds, middleware.DefaultRateLimitMaxRequests, middleware.DefaultRateLimitKeyPrefix)).Handle,
	}
}
