package svc

import (
	followrepo "go_zero-tiktok/app/communication/rpc/internal/dal/reposity"

	"gorm.io/gorm"
)

type Repositories struct {
	Follow *followrepo.UserFollowRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Follow: followrepo.NewUserFollowRepo(db),
	}
}
