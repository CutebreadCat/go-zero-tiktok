package reposity

import (
	"context"

	videointeractiontable "go_zero-tiktok/app/interaction/rpc/internal/dal/tables/video_interaction"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

// VideoInteractionRepo 视频交互（点赞/收藏）仓储
type VideoInteractionRepo struct {
	db *gorm.DB
}

func NewVideoInteractionRepo(db *gorm.DB) *VideoInteractionRepo {
	return &VideoInteractionRepo{db: db}
}

func (r *VideoInteractionRepo) LikeVideo(ctx context.Context, userID, videoID int64) error {
	if err := videointeractiontable.LikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.LikeVideo")
	}
	return nil
}

func (r *VideoInteractionRepo) CancelLikeVideo(ctx context.Context, userID, videoID int64) error {
	if err := videointeractiontable.CancelLikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.CancelLikeVideo")
	}
	return nil
}

func (r *VideoInteractionRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	videoIDs, total, err := videointeractiontable.GetLikedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoInteractionRepo.GetLikedVideoIDsByUserID")
	}
	return videoIDs, total, nil
}

func (r *VideoInteractionRepo) FavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := videointeractiontable.FavoriteVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.FavoriteVideo")
	}
	return nil
}

func (r *VideoInteractionRepo) CancelFavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := videointeractiontable.CancelFavoriteVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.CancelFavoriteVideo")
	}
	return nil
}

func (r *VideoInteractionRepo) GetFavoritedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	videoIDs, total, err := videointeractiontable.GetFavoritedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoInteractionRepo.GetFavoritedVideoIDsByUserID")
	}
	return videoIDs, total, nil
}

func (r *VideoInteractionRepo) GetLikeUserIDsByVideoID(ctx context.Context, videoID int64) ([]int64, error) {
	ids, err := videointeractiontable.GetLikeUserIDsByVideoID(ctx, r.db, videoID)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoInteractionRepo.GetLikeUserIDsByVideoID")
	}
	return ids, nil
}

func (r *VideoInteractionRepo) BatchAddLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error {
	if err := videointeractiontable.BatchAddLikeInteractions(ctx, r.db, videoID, userIDs); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.BatchAddLikeInteractions")
	}
	return nil
}

func (r *VideoInteractionRepo) BatchRemoveLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error {
	if err := videointeractiontable.BatchRemoveLikeInteractions(ctx, r.db, videoID, userIDs); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.BatchRemoveLikeInteractions")
	}
	return nil
}

func (r *VideoInteractionRepo) ApplyLikeEvent(ctx context.Context, action string, userID, videoID int64) error {
	if err := videointeractiontable.ApplyLikeEvent(ctx, r.db, action, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoInteractionRepo.ApplyLikeEvent")
	}
	return nil
}
