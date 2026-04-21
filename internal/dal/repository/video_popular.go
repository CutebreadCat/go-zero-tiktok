package repository

import (
	"context"
	"errors"

	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"
	"go_zero-tiktok/internal/types"

	goRedis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type VideoPopularRepo struct {
	db  *gorm.DB
	rdb *redis.Redis
}

func NewVideoPopularRepo(db *gorm.DB, rdb *redis.Redis) *VideoPopularRepo {
	return &VideoPopularRepo{db: db, rdb: rdb}
}

func (r *VideoPopularRepo) CreatePopularVideo(ctx context.Context, videoID string) error {
	return videopopulartable.CreatePopularVideo(ctx, r.db, videoID)
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	if err := videopopulartable.IncreaseVideoVisitCount(ctx, r.db, videoID, delta); err != nil {
		return err
	}
	if err := videopopulartable.IncrVideoVisitCountInRedis(ctx, r.rdb, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		logx.WithContext(ctx).Errorf("sync video visit count to redis failed: %v", err)
	}
	return nil
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	return videopopulartable.UpdateVideoLikeCount(ctx, r.db, videoID, delta)
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
	return videopopulartable.GetPopularVideoIDsByVisitCount(ctx, r.db, pageNum, pageSize)
}
