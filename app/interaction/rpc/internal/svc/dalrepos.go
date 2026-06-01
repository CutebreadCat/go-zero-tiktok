package svc

import (
	commentrepo "go_zero-tiktok/app/interaction/rpc/internal/dal/reposity"

	"gorm.io/gorm"
)

type Repositories struct {
	Comment *commentrepo.CommentRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Comment: commentrepo.NewCommentRepo(db),
	}
}
