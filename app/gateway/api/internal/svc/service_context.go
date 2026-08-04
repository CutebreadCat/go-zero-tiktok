// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	communicationservice "go_zero-tiktok/app/communication/rpc/communicationservice"
	"go_zero-tiktok/app/gateway/api/internal/config"
	"go_zero-tiktok/app/gateway/api/internal/middleware"
	interactionservice "go_zero-tiktok/app/interaction/rpc/interactionservice"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/app/video/rpc/videoservice"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config           config.Config
	UserRpc          userservice.UserService
	VideoRpc         videoservice.VideoService
	InteractionRpc   interactionservice.InteractionService
	CommunicationRpc communicationservice.CommunicationService
	RateLimit        rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:           c,
		UserRpc:          userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
		VideoRpc:         videoservice.NewVideoService(zrpc.MustNewClient(c.VideoRpc)),
		InteractionRpc:   interactionservice.NewInteractionService(zrpc.MustNewClient(c.InteractionRpc)),
		CommunicationRpc: communicationservice.NewCommunicationService(zrpc.MustNewClient(c.CommunicationRpc)),
		RateLimit:        middleware.NewRateLimitMiddleware().Handle,
	}
}
