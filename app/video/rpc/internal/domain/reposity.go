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
	GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error)
}

type IVideoLikerRepo interface {
	LikeVideo(ctx context.Context, userID, videoID int64) error
	CancelLikeVideo(ctx context.Context, userID, videoID int64) error
	GetLikedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error)
}

type StorageProvider interface {
	UploadFile(reader io.Reader, objectKey string) (string, error)
}
