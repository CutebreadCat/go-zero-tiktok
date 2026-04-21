package video_popular

import (
	"context"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	popularVideosRankKey = "popular_videos"
	popularVideosHashKey = "popular_videos:hash"
)

type PopularVideoWithHeat struct {
	HeatScore int64               `json:"heat_score"`
	Video     types.VideoBaseinfo `json:"video"`
}

func SetPopularVideoToRedis(ctx context.Context, rdb *redis.Redis, video types.VideoPopular) error {
	logger := logx.WithContext(ctx)

	if ok, err := rdb.Zadd(popularVideosRankKey, video.VisitCount, video.VideoID); !ok {
		logger.Errorf("set popular video to redis failed: %v", err)
		return xerr.New(1002, "设置热门视频到Redis失败")
	}

	return nil
}

func IncrVideoVisitCountInRedis(ctx context.Context, rdb *redis.Redis, videoID string) error {
	logger := logx.WithContext(ctx)

	if _, err := rdb.Zincrby(popularVideosRankKey, 1, videoID); err != nil {
		logger.Errorf("incr video visit count in redis failed: %v", err)
		return xerr.New(1002, "在zset中增加视频访问次数失败")
	}

	return nil
}

func GetVideoVisitCountFromRedis(ctx context.Context, rdb *redis.Redis, pageSize int, pageNum int) ([]string, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := int64((pageNum - 1) * pageSize)
	stop := int64(pageNum*pageSize - 1)
	pairs, err := rdb.ZrevrangeWithScores(popularVideosRankKey, start, stop)
	if err != nil {
		logger.Errorf("get video visit count from redis failed: %v", err)
		return nil, xerr.New(1002, "获取视频访问次数失败")
	}

	var result []string
	for _, pair := range pairs {
		videoId := pair.Key
		result = append(result, videoId)

	}

	return result, nil
}
