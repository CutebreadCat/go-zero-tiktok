// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"go_zero-tiktok/app/user/api/internal/config"
	"go_zero-tiktok/app/user/api/internal/middleware"
)

type ServiceContext struct {
	Config    config.Config
	RateLimit rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		RateLimit: middleware.NewRateLimitMiddleware().Handle,
	}
}
