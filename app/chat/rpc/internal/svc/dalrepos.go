package svc

import (
	chatrepo "go_zero-tiktok/app/chat/rpc/internal/dal/reposity"

	"gorm.io/gorm"
)

type Repositories struct {
	Chat *chatrepo.ChatRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Chat: chatrepo.NewChatRepo(db),
	}
}
