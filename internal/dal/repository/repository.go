package repository

import (
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	User       *UserBaseinfoRepo
	Video      *VideoBaseinfoRepo
	Popular    *VideoPopularRepo
	Comment    *CommentRepo
	VideoLiker *VideoLikerRepo
	UserFollow *UserFollowRepo
}

func NewRepositories(db *gorm.DB, rdb *redis.Redis) *Repositories {
	return &Repositories{
		User:       NewUserBaseinfoRepo(db),
		Video:      NewVideoBaseinfoRepo(db),
		Popular:    NewVideoPopularRepo(db, rdb),
		Comment:    NewCommentRepo(db),
		VideoLiker: NewVideoLikerRepo(db, rdb),
		UserFollow: NewUserFollowRepo(db),
	}
}

type repositoryInitializer struct{}

var _ = types.BaseResponse{}
