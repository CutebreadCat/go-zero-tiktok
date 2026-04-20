package redis

import (
	"context"

	videolikertable "go_zero-tiktok/internal/dal/tables/video_liker"
	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type VideoPopularRepo struct {
	rdb *redis.Redis
}

func NewVideoPopularRepo(rdb *redis.Redis) *VideoPopularRepo {
	return &VideoPopularRepo{rdb: rdb}
}

func (r *VideoPopularRepo) SetPopularVideoToRedis(ctx context.Context, video types.VideoBaseinfo, visitCount int64) error {
	return videopopulartable.SetPopularVideoToRedis(ctx, r.rdb, video, visitCount)
}

func (r *VideoPopularRepo) IncrVideoVisitCountInRedis(ctx context.Context, videoID string) error {
	return videopopulartable.IncrVideoVisitCountInRedis(ctx, r.rdb, videoID)
}

func (r *VideoPopularRepo) GetVideoVisitCountFromRedis(ctx context.Context, pageSize int, pageNum int) ([]string, error) {
	return videopopulartable.GetVideoVisitCountFromRedis(ctx, r.rdb, pageSize, pageNum)
}

func (r *VideoPopularRepo) GetVideoVisitCountByIDInHash(ctx context.Context, videoIDs []string) ([]map[string]string, error) {
	return videopopulartable.GetVideoVisitCountByIDInHash(ctx, r.rdb, videoIDs)
}

type VideoLikerRepo struct {
	rdb *redis.Redis
}

func NewVideoLikerRepo(rdb *redis.Redis) *VideoLikerRepo {
	return &VideoLikerRepo{rdb: rdb}
}

func (r *VideoLikerRepo) AddVideoLike(ctx context.Context, userID, videoID string) error {
	return videolikertable.AddVideoLike(ctx, r.rdb, userID, videoID)
}

func (r *VideoLikerRepo) RemoveVideoLike(ctx context.Context, userID, videoID string) error {
	return videolikertable.RemoveVideoLike(ctx, r.rdb, userID, videoID)
}

func (r *VideoLikerRepo) GetLikedVideoIDs(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	return videolikertable.GetLikedVideoIDs(ctx, r.rdb, userID, pageNumber, pageSize)
}

func (r *VideoLikerRepo) ResetLikedVideoIDs(ctx context.Context, userID string, videoIDs []string) error {
	return videolikertable.ResetLikedVideoIDs(ctx, r.rdb, userID, videoIDs)
}
