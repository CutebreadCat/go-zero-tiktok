// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb"
	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	"go_zero-tiktok/app/user/rpc/userservice"
	videopb "go_zero-tiktok/app/video/rpc/video_pb/video_pb"
	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/app/chat/domain/websocket"
	"go_zero-tiktok/internal/infra/ai"
	wscache "go_zero-tiktok/internal/infra/cache/ws"
	"go_zero-tiktok/internal/infra/storage/aliyun"
	"go_zero-tiktok/internal/middleware"
	"go_zero-tiktok/internal/middleware/government/breaker"
	"go_zero-tiktok/internal/middleware/government/limiter"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Cache  *wscache.RedisCache

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
	// 初始化 Redis
	rds := redis.MustNewRedis(config.Redis)
	c := wscache.NewRedisCache(rds)

	// 初始化阿里云配置
	aliyun.GetAliConfig()
	aliyun.AliInit()

	// 创建 RPC clients
	videoRpc := videopb.NewVideoServiceClient(zrpc.MustNewClient(config.VideoRpc).Conn())
	interactionRpc := interactionpb.NewInteractionServiceClient(zrpc.MustNewClient(config.InteractionRpc).Conn())
	communicationRpc := communicationpb.NewCommunicationServiceClient(zrpc.MustNewClient(config.CommunicationRpc).Conn())
	chatRpc := chatpb.NewChatServiceClient(zrpc.MustNewClient(config.ChatRpc).Conn())

	aiLimiter := limiter.New(rds, ai.DefaultLimitSeconds, ai.DefaultLimitMaxRequests, ai.DefaultLimitKeyPrefix)
	wsLimiter := limiter.New(rds, websocket.DefaultLimitSeconds, websocket.DefaultLimitMaxRequests, websocket.DefaultLimitKeyPrefix)
	aiBreaker := breaker.New()

	aiAgent, err := ai.NewAgent(context.Background(), aiLimiter, aiBreaker)
	if err != nil {
		logx.Must(err)
	}
	aiChat := websocket.NewAIChat(aiAgent, c)

	// 创建 Chat 仓储适配器（通过 gRPC 调用 chat 服务）
	chatRepoAdapter := NewChatRepoAdapter(chatRpc)

	// 创建 Hub
	hub := websocket.NewHub(c, c, c, chatRepoAdapter, chatRepoAdapter, aiChat, wsLimiter)

	// 初始化 MQ 并注入 writer
	mq := InitMQ(config.Kafka, hub, aiChat)

	return &ServiceContext{
		Config: config,

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
		RateLimit: middleware.NewRateLimitMiddleware(limiter.New(rds, middleware.DefaultRateLimitSeconds, middleware.DefaultRateLimitMaxRequests, middleware.DefaultRateLimitKeyPrefix)).Handle,
	}
}
