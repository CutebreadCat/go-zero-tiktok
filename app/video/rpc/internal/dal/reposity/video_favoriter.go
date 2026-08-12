package reposity

import (
	"context"

	videofavoritertable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_favoriter"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type VideoFavoriterRepo struct {
	db *gorm.DB
}

func NewVideoFavoriterRepo(db *gorm.DB) *VideoFavoriterRepo {
	return &VideoFavoriterRepo{db: db}
}

func (r *VideoFavoriterRepo) FavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := videofavoritertable.FavoriteVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoFavoriterRepo.FavoriteVideo")
	}
	return nil
}

func (r *VideoFavoriterRepo) CancelFavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := videofavoritertable.CancelFavoriteVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoFavoriterRepo.CancelFavoriteVideo")
	}
	return nil
}

func (r *VideoFavoriterRepo) GetFavoritedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	videoIDs, total, err := videofavoritertable.GetFavoritedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoFavoriterRepo.GetFavoritedVideoIDsByUserID")
	}
	return videoIDs, total, nil
}
