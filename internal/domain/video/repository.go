package video

import (
	"context"
	"go_zero-tiktok/internal/types"
)

type IVideoRepo interface {
	CreateVideoFromParams(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error
	SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error)
	GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
}

type IPopularRepo interface {
	CreatePopularVideo(ctx context.Context, videoID string) error
	IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error
	UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error
	GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error)
}

type IVideoLikerRepo interface {
	LikeVideo(ctx context.Context, userID, videoID string) error
	CancelLikeVideo(ctx context.Context, userID, videoID string) error
	GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error)
}
