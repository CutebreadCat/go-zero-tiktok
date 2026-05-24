package mock

import (
	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/svc"
)

func NewServiceContext(
	userRepo *UserRepo,
	videoRepo *VideoRepo,
	popularRepo *PopularRepo,
	commentRepo *CommentRepo,
	videoLikerRepo *VideoLikerRepo,
	userFollowRepo *UserFollowRepo,
	chatRepo *ChatRepo,
) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			Auth: config.AuthConfig{
				AccessSecret: "test-secret",
			},
		},
		Dal: &svc.Repositories{
			User:       userRepo,
			Video:      videoRepo,
			Popular:    popularRepo,
			Comment:    commentRepo,
			VideoLiker: videoLikerRepo,
			UserFollow: userFollowRepo,
			Chat:       chatRepo,
		},
	}
}
