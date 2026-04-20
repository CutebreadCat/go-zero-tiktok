package video_popular

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	goRedis "github.com/redis/go-redis/v9"
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

func GetVideoVisitCountByIDInHash(ctx context.Context, rdb *redis.Redis, videoIDs []string) ([]map[string]string, error) {
	logger := logx.WithContext(ctx)

	if len(videoIDs) == 0 {
		return nil, nil
	}

	var videos []map[string]string
	for _, videoID := range videoIDs {
		popularVideoHashKey := popularVideosHashKey + ":" + videoID
		visitCountStr, err := rdb.Hgetall(popularVideoHashKey)
		if err != nil {
			if errors.Is(err, goRedis.Nil) {
				logger.Errorf("video %s not found in Redis, skipping.", videoID)
				return nil, goRedis.Nil
			}
			logger.Errorf("get video visit count from redis failed: %v", err)
			return nil, xerr.New(1002, "获取视频访问次数失败")
		}
		videos = append(videos, visitCountStr)
	}

	return videos, nil
}

func SetPopularVideoToRedis(ctx context.Context, rdb *redis.Redis, video types.VideoBaseinfo, visitCount int64) error {
	logger := logx.WithContext(ctx)

	if visitCount < 0 {
		visitCount = 0
	}

	if ok, err := rdb.Zadd(popularVideosRankKey, visitCount, video.VideoID); !ok {
		logger.Errorf("set popular video to redis failed: %v", err)
		return xerr.New(1002, "设置热门视频到Redis失败")
	}

	videoinfo := make(map[string]string)
	videoinfo["video_id"] = video.VideoID
	videoinfo["author_id"] = video.AuthorID
	videoinfo["video_url"] = video.VideoURL
	videoinfo["cover_url"] = video.CoverURL
	videoinfo["title"] = video.Title
	videoinfo["description"] = video.Description
	videoinfo["visit_count"] = strconv.FormatInt(video.VisitCount, 10)
	videoinfo["like_count"] = strconv.FormatInt(video.LikeCount, 10)
	videoinfo["comment_count"] = strconv.FormatInt(video.CommentCount, 10)

	popularVideoHashKey := popularVideosHashKey + ":" + video.VideoID
	if err := rdb.Hmset(popularVideoHashKey, videoinfo); err != nil {
		logger.Errorf("set popular video hash failed: %v", err)
		return xerr.New(1002, "设置热门视频哈希失败")
	}

	return nil
}

func IncrVideoVisitCountInRedis(ctx context.Context, rdb *redis.Redis, videoID string) error {
	logger := logx.WithContext(ctx)

	if _, err := rdb.Zincrby(popularVideosRankKey, 1, videoID); err != nil {
		logger.Errorf("incr video visit count in redis failed: %v", err)
		return xerr.New(1002, "在zset中增加视频访问次数失败")
	}

	popularVideoHashKey := popularVideosHashKey + ":" + videoID
	if _, err := rdb.Hincrby(popularVideoHashKey, "visit_count", 1); err != nil {
		logger.Errorf("incr video visit count hash failed: %v", err)
		return xerr.New(1002, "在hash中增加视频访问次数失败")
	}
	return nil
}

func GetVideoVisitCountFromRedis(ctx context.Context, rdb *redis.Redis, pageSize int, pageNum int) ([]PopularVideoWithHeat, error) {
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

	result := make([]PopularVideoWithHeat, 0, len(pairs))
	for _, pair := range pairs {
		videoJSON, err := rdb.Hget(popularVideosHashKey, pair.Key)
		if err != nil {
			logger.Errorf("get video hash from redis failed: %v", err)
			return nil, xerr.New(1002, "获取视频哈希失败")
		}

		var video types.VideoBaseinfo
		if err := json.Unmarshal([]byte(videoJSON), &video); err != nil {
			logger.Errorf("unmarshal video hash failed: %v", err)
			return nil, xerr.New(1002, "反序列化视频哈希失败")
		}

		result = append(result, PopularVideoWithHeat{HeatScore: pair.Score, Video: video})
	}

	return result, nil
}
