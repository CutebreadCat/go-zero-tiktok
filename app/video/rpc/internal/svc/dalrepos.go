package svc

import (
	videorepo "go_zero-tiktok/app/video/rpc/internal/dal/reposity"

	"gorm.io/gorm"
)

type Repositories struct {
	Video      *videorepo.VideoBaseinfoRepo
	Popular    *videorepo.VideoPopularRepo
	VideoLiker *videorepo.VideoLikerRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Video:      videorepo.NewVideoBaseinfoRepo(db),
		Popular:    videorepo.NewVideoPopularRepo(db),
		VideoLiker: videorepo.NewVideoLikerRepo(db),
	}
}
