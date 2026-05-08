// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/dal"
	repository "go_zero-tiktok/internal/dal/repository"
	"go_zero-tiktok/internal/infra/storage/aliyun"
	"go_zero-tiktok/internal/websocket"

	"go_zero-tiktok/internal/cache"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB

	Cache *cache.RedisCache
	Rdb   *redis.Redis
	Dal   *repository.Repositories
	Hub   *websocket.Hub
}

func NewServiceContext(config config.Config) *ServiceContext {
	logx.Must(dal.InitMysql(config.DataSource))
	dal.InitRedis(config.Redis)

	// 初始化阿里云配置
	aliyun.GetAliConfig()
	aliyun.AliInit()

	c := cache.NewRedisCache(dal.Rdb)
	dalRepo := repository.NewRepositories(dal.Db, dal.Rdb)

	return &ServiceContext{
		Config: config,
		DB:     dal.Db,
		Rdb:    dal.Rdb,
		Dal:    dalRepo,
		Cache:  c,
		Hub:    websocket.NewHub(c, dalRepo.Chat),
	}
}
