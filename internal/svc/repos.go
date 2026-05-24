package svc

import (
	repository "go_zero-tiktok/internal/dal/repository"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	User       IUserRepo
	Video      IVideoRepo
	Popular    IPopularRepo
	Comment    ICommentRepo
	VideoLiker IVideoLikerRepo
	UserFollow IUserFollowRepo
	Chat       IChatRepo
}

func NewRepositories(db *gorm.DB, rdb *redis.Redis) *Repositories {
	return &Repositories{
		User:       repository.NewUserBaseinfoRepo(db),
		Video:      repository.NewVideoBaseinfoRepo(db),
		Popular:    repository.NewVideoPopularRepo(db, rdb),
		Comment:    repository.NewCommentRepo(db),
		VideoLiker: repository.NewVideoLikerRepo(db, rdb),
		UserFollow: repository.NewUserFollowRepo(db),
		Chat:       repository.NewChatRepo(db),
	}
}
