package svc

import (
	videorepo "go_zero-tiktok/app/video/rpc/internal/dal/reposity"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	Video     *videorepo.VideoBaseinfoRepo
	VideoStat *videorepo.VideoStatRepo
	Feed      *videorepo.FeedRepo
}

func NewRepositories(db *gorm.DB, rdb *redis.Redis) (*Repositories, error) {
	videoRepo, err := videorepo.NewVideoBaseinfoRepo(db)
	if err != nil {
		return nil, err
	}
	return &Repositories{
		Video:     videoRepo,
		VideoStat: videorepo.NewVideoStatRepo(db),
		Feed:      videorepo.NewFeedRepo(rdb),
	}, nil
}
