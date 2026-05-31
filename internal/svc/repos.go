package svc

import (
	repository "go_zero-tiktok/internal/dal/repository"
	chatdomain "go_zero-tiktok/internal/domain/chat"
	commentdomain "go_zero-tiktok/internal/domain/comment"
	userdomain "go_zero-tiktok/internal/domain/user"
	videodomain "go_zero-tiktok/internal/domain/video"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	User       userdomain.IUserRepo
	Video      videodomain.IVideoRepo
	Popular    videodomain.IPopularRepo
	Comment    commentdomain.ICommentRepo
	VideoLiker videodomain.IVideoLikerRepo
	UserFollow userdomain.IUserFollowRepo
	Chat       chatdomain.IChatRepo
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
