package repository

import (
	"context"
	"errors"
	"fmt"

	videolikertable "go_zero-tiktok/internal/dal/tables/video_liker"

	goRedis "github.com/redis/go-redis/v9"
	pkgerrors "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type VideoLikerRepo struct {
	db  *gorm.DB
	rdb *redis.Redis
}

func NewVideoLikerRepo(db *gorm.DB, rdb *redis.Redis) *VideoLikerRepo {
	return &VideoLikerRepo{db: db, rdb: rdb}
}

func (r *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	if err := videolikertable.LikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoLikerRepo.LikeVideo")
	}
	if err := videolikertable.AddVideoLike(ctx, r.rdb, userID, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		fmt.Printf("sync like to redis failed: %v\n", err)
	}
	return nil
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if err := videolikertable.CancelLikeVideo(ctx, r.db, userID, videoID); err != nil {
		return pkgerrors.WithMessage(err, "VideoLikerRepo.CancelLikeVideo")
	}
	if err := videolikertable.RemoveVideoLike(ctx, r.rdb, userID, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		fmt.Printf("sync unlike to redis failed: %v\n", err)
	}
	return nil
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	videoIDs, total, err := videolikertable.GetLikedVideoIDs(ctx, r.rdb, userID, pageNumber, pageSize)
	if err == nil && total > 0 {
		return videoIDs, total, nil
	}
	if err != nil && !errors.Is(err, goRedis.Nil) {
		fmt.Printf("read liked videos from redis failed: %v\n", err)
	}

	videoIDs, total, err = videolikertable.GetLikedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoLikerRepo.GetLikedVideoIDsByUserID")
	}

	allLikedVideoIDs, listErr := videolikertable.GetAllLikedVideoIDsByUserID(ctx, r.db, userID)
	if listErr != nil {
		fmt.Printf("query all liked videos for cache backfill failed: %v\n", listErr)
		return videoIDs, total, nil
	}

	if cacheErr := videolikertable.ResetLikedVideoIDs(ctx, r.rdb, userID, allLikedVideoIDs); cacheErr != nil {
		fmt.Printf("backfill liked videos cache failed: %v\n", cacheErr)
	}

	return videoIDs, total, nil
}
