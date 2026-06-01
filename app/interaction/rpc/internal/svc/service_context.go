package svc

import (
	"go_zero-tiktok/app/interaction/rpc/internal/config"
	commentdomain "go_zero-tiktok/app/interaction/rpc/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	DB             *gorm.DB
	Dal            *Repositories
	CommentService *commentdomain.CommentService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	dalRepo := NewRepositories(db)
	videoVisitAdapter := NewVideoVisitAdapter(db)

	return &ServiceContext{
		Config:         c,
		DB:             db,
		Dal:            dalRepo,
		CommentService: commentdomain.NewCommentService(dalRepo.Comment, videoVisitAdapter),
	}
}
