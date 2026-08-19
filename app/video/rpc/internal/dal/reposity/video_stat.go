package reposity

import (
	"context"

	videostattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/pkg/contract"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type VideoStatRepo struct {
	db *gorm.DB
}

func NewVideoStatRepo(db *gorm.DB) *VideoStatRepo {
	return &VideoStatRepo{db: db}
}

func (r *VideoStatRepo) CreatePopularVideo(ctx context.Context, videoID int64) error {
	if err := videostattable.CreatePopularVideo(ctx, r.db, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoStatRepo.CreatePopularVideo")
	}
	return nil
}

func (r *VideoStatRepo) IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error {
	if err := videostattable.IncreaseVideoVisitCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoStatRepo.IncreaseVideoVisitCount")
	}
	return nil
}

func (r *VideoStatRepo) UpdateVideoLikeCount(ctx context.Context, videoID int64, delta int64) error {
	if err := videostattable.UpdateVideoLikeCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoStatRepo.UpdateVideoLikeCount")
	}
	return nil
}

func (r *VideoStatRepo) UpdateVideoFavoriteCount(ctx context.Context, videoID int64, delta int64) error {
	if err := videostattable.UpdateVideoFavoriteCount(ctx, r.db, videoID, delta); err != nil {
		return pkgerrors.WithMessage(err, "VideoStatRepo.UpdateVideoFavoriteCount")
	}
	return nil
}

func (r *VideoStatRepo) SetLikeCount(ctx context.Context, videoID int64, count int64) error {
	if err := videostattable.SetLikeCount(ctx, r.db, videoID, count); err != nil {
		return pkgerrors.WithMessage(err, "VideoStatRepo.SetLikeCount")
	}
	return nil
}

func (r *VideoStatRepo) GetLikeCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error) {
	counts, err := videostattable.GetLikeCounts(ctx, r.db, videoIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoStatRepo.GetLikeCounts")
	}
	return counts, nil
}

func (r *VideoStatRepo) GetPopularVideosByIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoPopular, error) {
	rows, err := videostattable.GetPopularVideosByIDs(ctx, r.db, videoIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoStatRepo.GetPopularVideosByIDs")
	}

	result := make(map[int64]types.VideoPopular, len(videoIDs))
	for _, p := range rows {
		result[p.VideoID] = r.VideoStatToResponse(&p)
	}
	for _, id := range videoIDs {
		if _, ok := result[id]; !ok {
			result[id] = types.VideoPopular{VideoID: id}
		}
	}
	return result, nil
}

func (r *VideoStatRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
	rows, total, err := videostattable.GetPopularVideoIDsByVisitCount(ctx, r.db, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoStatRepo.GetPopularVideoIDsByVisitCount")
	}
	return r.VideoStatsToResponse(rows), total, nil
}

func (r *VideoStatRepo) GetPopularVideosByCursor(ctx context.Context, score, videoID int64, limit int32) ([]types.VideoPopular, error) {
	rows, err := videostattable.GetPopularVideosByCursor(ctx, r.db, score, videoID, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoStatRepo.GetPopularVideosByCursor")
	}
	return r.VideoStatsToResponse(rows), nil
}

func (r *VideoStatRepo) VideoStatToResponse(popular *videostattable.VideoStat) types.VideoPopular {
	return types.VideoPopular{
		VideoID:       popular.VideoID,
		VisitCount:    popular.VisitCount,
		LikeCount:     popular.LikeCount,
		CommentCount:  popular.CommentCount,
		FavoriteCount: popular.FavoriteCount,
	}
}

func (r *VideoStatRepo) VideoStatsToResponse(populars []videostattable.VideoStat) []types.VideoPopular {
	result := make([]types.VideoPopular, 0, len(populars))
	for _, p := range populars {
		result = append(result, r.VideoStatToResponse(&p))
	}
	return result
}
