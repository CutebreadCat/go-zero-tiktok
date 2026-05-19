package video_popular

import (
	"context"

	"go_zero-tiktok/internal/svc/xerr"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	popularVideosRankKey = "popular_videos"
	popularVideosHashKey = "popular_videos:hash"
)

func SetPopularVideoToRedis(ctx context.Context, rdb *redis.Redis, video VideoPopular) error {
	if ok, err := rdb.Zadd(popularVideosRankKey, video.VisitCount, video.VideoID); !ok {
		return xerr.Wrap(err, "set popular video to redis failed")
	}

	return nil
}

func IncrVideoVisitCountInRedis(ctx context.Context, rdb *redis.Redis, videoID string) error {
	if _, err := rdb.Zincrby(popularVideosRankKey, 1, videoID); err != nil {
		return xerr.Wrap(err, "incr video visit count in redis failed")
	}

	return nil
}

func GetVideoVisitCountFromRedis(ctx context.Context, rdb *redis.Redis, pageSize int, pageNum int) ([]string, error) {
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
		return nil, xerr.Wrap(err, "get video visit count from redis failed")
	}

	var result []string
	for _, pair := range pairs {
		videoId := pair.Key
		result = append(result, videoId)
	}

	return result, nil
}
