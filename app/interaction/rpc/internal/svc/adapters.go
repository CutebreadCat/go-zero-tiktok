package svc

import (
	"context"

	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"

	"gorm.io/gorm"
)

type VideoVisitAdapter struct {
	db *gorm.DB
}

func NewVideoVisitAdapter(db *gorm.DB) *VideoVisitAdapter {
	return &VideoVisitAdapter{db: db}
}

func (a *VideoVisitAdapter) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	return videopopulartable.IncreaseVideoVisitCount(ctx, a.db, videoID, delta)
}
