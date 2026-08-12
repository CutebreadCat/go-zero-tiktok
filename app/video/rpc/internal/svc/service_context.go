package svc

import (
	"go_zero-tiktok/app/video/rpc/internal/config"
	videodomain "go_zero-tiktok/app/video/rpc/internal/domain"
	"go_zero-tiktok/pkg/storage/aliyun"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	Dal          *Repositories
	VideoService *videodomain.VideoService
	Storage      *StorageAdapter
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	// 初始化阿里云配置
	aliyun.LoadConfig()
	aliyun.InitClient()

	dalRepo := NewRepositories(db)
	storageAdapter := &StorageAdapter{}

	return &ServiceContext{
		Config:       c,
		DB:           db,
		Dal:          dalRepo,
		VideoService: videodomain.NewVideoService(dalRepo.Video, dalRepo.VideoStat, dalRepo.VideoLiker, dalRepo.VideoFavoriter, storageAdapter),
		Storage:      storageAdapter,
	}
}
