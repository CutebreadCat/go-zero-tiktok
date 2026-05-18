package repository

import (
	"context"
	"errors"
	"fmt"

	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"
	"go_zero-tiktok/internal/types"

	goRedis "github.com/redis/go-redis/v9"
	pkgerrors "github.com/pkg/errors"
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
	if err := videopopulartable.CreatePopularVideo(ctx, r.db, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.CreatePopularVideo")
	}
	return nil
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	if err := videopopulartable.IncreaseVideoVisitCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.IncreaseVideoVisitCount")
	}
	if err := videopopulartable.IncrVideoVisitCountInRedis(ctx, r.rdb, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		fmt.Printf("sync video visit count to redis failed: %v\n", err)
	}
	return nil
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	if err := videopopulartable.UpdateVideoLikeCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.UpdateVideoLikeCount")
	}
	return nil
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]videopopulartable.VideoPopular, int64, error) {
	rows, total, err := videopopulartable.GetPopularVideoIDsByVisitCount(ctx, r.db, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoPopularRepo.GetPopularVideoIDsByVisitCount")
	}
	return rows, total, nil
}

func (r *VideoPopularRepo) VideoPopularToResponse(popular *videopopulartable.VideoPopular) types.VideoPopular {
	return types.VideoPopular{
		VideoID:      popular.VideoID,
		VisitCount:   popular.VisitCount,
		LikeCount:    popular.LikeCount,
		CommentCount: popular.CommentCount,
	}
}

func (r *VideoPopularRepo) VideoPopularsToResponse(populars []videopopulartable.VideoPopular) []types.VideoPopular {
	result := make([]types.VideoPopular, 0, len(populars))
	for _, p := range populars {
		result = append(result, r.VideoPopularToResponse(&p))
	}
	return result
}
