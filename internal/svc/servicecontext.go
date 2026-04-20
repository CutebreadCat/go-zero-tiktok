// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/mw/ali"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Rdb    *redis.Redis
	Dal    *dal.Repositories
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	// 初始化阿里云配置
	ali.GetAliConfig()
	ali.AliInit()

	return &ServiceContext{
		Config: config,
		DB:     dal.Db,
		Rdb:    dal.Rdb,
		Dal:    dal.NewRepositories(dal.Db, dal.Rdb),
	}
}
