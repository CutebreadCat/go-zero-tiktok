package svc

import (
	videorepo "go_zero-tiktok/app/video/rpc/internal/dal/reposity"

	"gorm.io/gorm"
)

type Repositories struct {
	Video     *videorepo.VideoBaseinfoRepo
	VideoStat *videorepo.VideoStatRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Video:     videorepo.NewVideoBaseinfoRepo(db),
		VideoStat: videorepo.NewVideoStatRepo(db),
	}
}
