package reposity

import (
	"context"

	videolikertable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_liker"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type VideoLikerRepo struct {
	db *gorm.DB
}

func NewVideoLikerRepo(db *gorm.DB) *VideoLikerRepo {
	return &VideoLikerRepo{db: db}
}

func (r *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	if err := videolikertable.LikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoLikerRepo.LikeVideo")
	}
	return nil
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if err := videolikertable.CancelLikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoLikerRepo.CancelLikeVideo")
	}
	return nil
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	videoIDs, total, err := videolikertable.GetLikedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoLikerRepo.GetLikedVideoIDsByUserID")
	}
	return videoIDs, total, nil
}
