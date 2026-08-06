package svc

import (
	"go_zero-tiktok/app/communication/rpc/internal/config"
	followdomain "go_zero-tiktok/app/communication/rpc/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config            config.Config
	DB                *gorm.DB
	Dal               *Repositories
	UserFollowService *followdomain.UserFollowService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	dalRepo := NewRepositories(db)
	userRepoAdapter := NewUserRepoAdapter(c)

	return &ServiceContext{
		Config:            c,
		DB:                db,
		Dal:               dalRepo,
		UserFollowService: followdomain.NewUserFollowService(dalRepo.Follow, userRepoAdapter),
	}
}
