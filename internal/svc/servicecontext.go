// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Rdb    *redis.Redis
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	return &ServiceContext{
		Config: config,
		DB:     dal.Db,
		Rdb:    dal.Rdb,
	}
}
