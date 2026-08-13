package domain

import (
	"context"
	"go_zero-tiktok/pkg/contract"
	"io"
)

type IVideoRepo interface {
	CreateVideoFromParams(ctx context.Context, videoID, authorID int64, videoURL, coverURL, title, description string) error
	SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error)
	GetVideosByAuthorID(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
}

type IPopularRepo interface {
	CreatePopularVideo(ctx context.Context, videoID int64) error
	IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error
	UpdateVideoLikeCount(ctx context.Context, videoID int64, delta int64) error
	UpdateVideoFavoriteCount(ctx context.Context, videoID int64, delta int64) error
	GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error)
	// SetLikeCount 直接设置视频 like_count（供 syncer 以 Redis 为基准对齐 MySQL）。
	SetLikeCount(ctx context.Context, videoID int64, count int64) error
	// GetLikeCounts 批量查询视频 like_count，用于 Redis 未命中时回源。
	GetLikeCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error)
}

type IVideoInteractionRepo interface {
	LikeVideo(ctx context.Context, userID, videoID int64) error
	CancelLikeVideo(ctx context.Context, userID, videoID int64) error
	GetLikedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error)
	FavoriteVideo(ctx context.Context, userID, videoID int64) error
	CancelFavoriteVideo(ctx context.Context, userID, videoID int64) error
	GetFavoritedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error)
	// 以下供后台 syncer 批量同步 Redis 关系到 MySQL 使用。
	GetLikeUserIDsByVideoID(ctx context.Context, videoID int64) ([]int64, error)
	BatchAddLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error
	BatchRemoveLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error
	// ApplyLikeEvent 事务内应用点赞/取消点赞事件（Kafka 消费者落库用）。
	// action 取值与 interaction.LikeAction 一致："like" / "cancel"。
	ApplyLikeEvent(ctx context.Context, action string, userID, videoID int64) error
}

type StorageProvider interface {
	UploadFile(reader io.Reader, objectKey string) (string, error)
}
