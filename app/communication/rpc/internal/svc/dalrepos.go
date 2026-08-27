package svc

import (
	"go_zero-tiktok/app/communication/rpc/internal/dal/reposity"
	"gorm.io/gorm"
)

type Repositories struct {
	Follow  *reposity.UserFollowRepo
	Message *reposity.MessageRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Follow:  reposity.NewUserFollowRepo(db),
		Message: reposity.NewMessageRepo(db),
	}
}
