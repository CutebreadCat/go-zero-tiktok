package svc

import (
	"go_zero-tiktok/app/chat/rpc/internal/config"
	chatdomain "go_zero-tiktok/app/chat/rpc/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	Dal         *Repositories
	ChatService *chatdomain.ChatService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	dalRepo := NewRepositories(db)

	return &ServiceContext{
		Config:      c,
		DB:          db,
		Dal:         dalRepo,
		ChatService: chatdomain.NewChatService(dalRepo.Chat),
	}
}
