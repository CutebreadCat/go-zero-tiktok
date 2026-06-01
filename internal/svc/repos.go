package svc

import (
	repository "go_zero-tiktok/internal/dal/repository"
	userdomain "go_zero-tiktok/internal/domain/user"

	"gorm.io/gorm"
)

type Repositories struct {
	User userdomain.IUserRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User: repository.NewUserBaseinfoRepo(db),
	}
}
