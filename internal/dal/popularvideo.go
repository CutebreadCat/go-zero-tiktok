package dal

import (
	"context"
	"encoding/json"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	popularVideosRankKey = "popular_videos"
	popularVideosHashKey = "popular_videos:hash"
)

type PopularVideoWithHeat struct {
	HeatScore int64               `json:"heat_score"`
	Video     types.VideoBaseinfo `json:"video"`
}

func CreatePopularVideo(ctx context.Context, videoID string) error {
	logger := logx.WithContext(ctx)

	record := &types.VideoPopular{
		VideoID:      videoID,
		VisitCount:   0,
		LikeCount:    0,
		CommentCount: 0,
	}

	if err := Db.WithContext(ctx).Create(record).Error; err != nil {
		logger.Errorf("create popular video failed: %v", err)
		return err
	}

	return nil
}

func IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	logger := logx.WithContext(ctx)

	if delta <= 0 {
		delta = 1
	}

	result := Db.WithContext(ctx).
		Model(&types.VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + ?", delta))
	if result.Error != nil {
		logger.Errorf("increase video visit count failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Errorf("increase video visit count failed: %v", gorm.ErrRecordNotFound)
		return gorm.ErrRecordNotFound
	}

	return nil
}

func UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).
		Model(&types.VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("like_count", gorm.Expr("CASE WHEN like_count + ? < 0 THEN 0 ELSE like_count + ? END", delta, delta))
	if result.Error != nil {
		logger.Errorf("update video like count failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Errorf("update video like count failed: %v", gorm.ErrRecordNotFound)
		return gorm.ErrRecordNotFound
	}

	return nil
}

func GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]string, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.VideoPopular{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get popular video count failed: %v", err)
		return nil, 0, err
	}

	var rows []types.VideoPopular
	offset := (pageNum - 1) * pageSize
	if err := query.Order("visit_count DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		logger.Errorf("get popular videos failed: %v", err)
		return nil, 0, err
	}

	videoIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

func SetPopularVideoToRedis(ctx context.Context, video types.VideoBaseinfo, visitCount int64) error {
	logger := logx.WithContext(ctx)

	if visitCount < 0 {
		visitCount = 0
	}

	if ok, err := Rdb.Zadd(popularVideosRankKey, visitCount, video.VideoID); !ok {
		logger.Errorf("set popular video to redis failed: %v", err)
		return err
	}

	videoJSON, err := json.Marshal(video)
	if err != nil {
		logger.Errorf("marshal popular video failed: %v", err)
		return err
	}

	if err := Rdb.Hset(popularVideosHashKey, video.VideoID, string(videoJSON)); err != nil {
		logger.Errorf("set popular video hash failed: %v", err)
		return err
	}

	return nil
}

func IncrVideoVisitCountInRedis(ctx context.Context, videoID string) error {
	logger := logx.WithContext(ctx)

	if _, err := Rdb.Zincrby(popularVideosRankKey, 1, videoID); err != nil {
		logger.Errorf("incr video visit count in redis failed: %v", err)
		return err
	}

	return nil
}

func GetVideoVisitCountFromRedis(ctx context.Context, pageSize int, pageNum int) ([]PopularVideoWithHeat, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := int64((pageNum - 1) * pageSize)
	stop := int64(pageNum*pageSize - 1)
	pairs, err := Rdb.ZrevrangeWithScores(popularVideosRankKey, start, stop)
	if err != nil {
		logger.Errorf("get video visit count from redis failed: %v", err)
		return nil, err
	}

	result := make([]PopularVideoWithHeat, 0, len(pairs))
	for _, pair := range pairs {
		videoJSON, err := Rdb.Hget(popularVideosHashKey, pair.Key)
		if err != nil {
			logger.Errorf("get video hash from redis failed: %v", err)
			return nil, err
		}

		var video types.VideoBaseinfo
		if err := json.Unmarshal([]byte(videoJSON), &video); err != nil {
			logger.Errorf("unmarshal video hash failed: %v", err)
			return nil, err
		}

		result = append(result, PopularVideoWithHeat{
			HeatScore: pair.Score,
			Video:     video,
		})
	}

	return result, nil

}
