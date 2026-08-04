package reposity

import (
	"context"

	videopopulartable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_popular"
	"go_zero-tiktok/pkg/contract"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type VideoPopularRepo struct {
	db *gorm.DB
}

func NewVideoPopularRepo(db *gorm.DB) *VideoPopularRepo {
	return &VideoPopularRepo{db: db}
}

func (r *VideoPopularRepo) CreatePopularVideo(ctx context.Context, videoID int64) error {
	if err := videopopulartable.CreatePopularVideo(ctx, r.db, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.CreatePopularVideo")
	}
	return nil
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error {
	if err := videopopulartable.IncreaseVideoVisitCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.IncreaseVideoVisitCount")
	}
	return nil
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID int64, delta int64) error {
	if err := videopopulartable.UpdateVideoLikeCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoPopularRepo.UpdateVideoLikeCount")
	}
	return nil
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
	rows, total, err := videopopulartable.GetPopularVideoIDsByVisitCount(ctx, r.db, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoPopularRepo.GetPopularVideoIDsByVisitCount")
	}
	return r.VideoPopularsToResponse(rows), total, nil
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
