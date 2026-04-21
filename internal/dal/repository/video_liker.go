package repository

import (
	"context"
	"errors"

	videolikertable "go_zero-tiktok/internal/dal/tables/video_liker"

	goRedis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
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
		return err
	}
	if err := videolikertable.AddVideoLike(ctx, r.rdb, userID, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		logx.WithContext(ctx).Errorf("sync like to redis failed: %v", err)
	}
	return nil
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if err := videolikertable.CancelLikeVideo(ctx, r.db, userID, videoID); err != nil {
		return err
	}
	if err := videolikertable.RemoveVideoLike(ctx, r.rdb, userID, videoID); err != nil && !errors.Is(err, goRedis.Nil) {
		logx.WithContext(ctx).Errorf("sync unlike to redis failed: %v", err)
	}
	return nil
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	videoIDs, total, err := videolikertable.GetLikedVideoIDs(ctx, r.rdb, userID, pageNumber, pageSize)
	if err == nil && total > 0 {
		return videoIDs, total, nil
	}
	if err != nil && !errors.Is(err, goRedis.Nil) {
		logx.WithContext(ctx).Errorf("read liked videos from redis failed: %v", err)
	}

	videoIDs, total, err = videolikertable.GetLikedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, err
	}

	allLikedVideoIDs, listErr := videolikertable.GetAllLikedVideoIDsByUserID(ctx, r.db, userID)
	if listErr != nil {
		logx.WithContext(ctx).Errorf("query all liked videos for cache backfill failed: %v", listErr)
		return videoIDs, total, nil
	}

	if cacheErr := videolikertable.ResetLikedVideoIDs(ctx, r.rdb, userID, allLikedVideoIDs); cacheErr != nil {
		logx.WithContext(ctx).Errorf("backfill liked videos cache failed: %v", cacheErr)
	}

	return videoIDs, total, nil
}
