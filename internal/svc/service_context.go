// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"
	chatdomain "go_zero-tiktok/internal/domain/chat"
	commentdomain "go_zero-tiktok/internal/domain/comment"
	userdomain "go_zero-tiktok/internal/domain/user"
	userfollowdomain "go_zero-tiktok/internal/domain/userfollow"
	videodomain "go_zero-tiktok/internal/domain/video"
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
	Config            config.Config
	DB                *gorm.DB
	Cache             *wscache.RedisCache
	Rdb               *redis.Redis
	Dal               *Repositories
	VideoService      *videodomain.VideoService
	CommentService    *commentdomain.CommentService
	UserFollowService *userfollowdomain.UserFollowService
	ChatService       *chatdomain.ChatService
	UserAuthService   *userdomain.AuthService
	UserMfaService    *userdomain.MfaService
	UserProfileService *userdomain.ProfileService
	UserJwchService   *userdomain.JwchService
	Hub               *websocket.Hub
	AIChat            *websocket.AIChat
	MQ                *MQComponents
	RateLimit         rest.Middleware
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	// 初始化阿里云配置
	aliyun.GetAliConfig()
	aliyun.AliInit()

	c := wscache.NewRedisCache(dal.Rdb)
	dalRepo := NewRepositories(dal.Db, dal.Rdb)

	// 创建适配器
	tokenAdapter := &TokenAdapter{}
	mfaAdapter := &MfaAdapter{}
	storageAdapter := &StorageAdapter{}
	jwchFactory := &JwchClientFactoryAdapter{}

	// 组装 domain service
	videoSvc := videodomain.NewVideoService(dalRepo.Video, dalRepo.Popular, dalRepo.VideoLiker)
	commentSvc := commentdomain.NewCommentService(dalRepo.Comment, dalRepo.Popular)
	userFollowSvc := userfollowdomain.NewUserFollowService(dalRepo.UserFollow, dalRepo.User)
	chatSvc := chatdomain.NewChatService(dalRepo.Chat)
	userAuthSvc := userdomain.NewAuthService(dalRepo.User, tokenAdapter, mfaAdapter, config.Auth.AccessSecret, dal.Rdb)
	userMfaSvc := userdomain.NewMfaService(dalRepo.User, mfaAdapter, mfaAdapter)
	userProfileSvc := userdomain.NewProfileService(dalRepo.User, storageAdapter)
	userJwchSvc := userdomain.NewJwchService(dalRepo.User, jwchFactory)

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
		Config:             config,
		DB:                 dal.Db,
		Rdb:                dal.Rdb,
		Dal:                dalRepo,
		VideoService:       videoSvc,
		CommentService:     commentSvc,
		UserFollowService:  userFollowSvc,
		ChatService:        chatSvc,
		UserAuthService:    userAuthSvc,
		UserMfaService:     userMfaSvc,
		UserProfileService: userProfileSvc,
		UserJwchService:    userJwchSvc,
		Cache:              c,
		Hub:                hub,
		AIChat:             aiChat,
		MQ:                 mq,
		RateLimit:          middleware.NewRateLimitMiddleware(limiter.New(dal.Rdb, middleware.DefaultRateLimitSeconds, middleware.DefaultRateLimitMaxRequests, middleware.DefaultRateLimitKeyPrefix)).Handle,
	}
}
