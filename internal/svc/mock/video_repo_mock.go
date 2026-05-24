package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type VideoRepo struct {
	CreateVideoFromParamsFn   func(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error
	SearchVideosByKeywordFn   func(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideosByIDsFn          func(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error)
	GetVideosByAuthorIDFn     func(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
}

func (m *VideoRepo) CreateVideoFromParams(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error {
	if m.CreateVideoFromParamsFn != nil {
		return m.CreateVideoFromParamsFn(ctx, videoID, authorID, videoURL, coverURL, title, description)
	}
	return nil
}

func (m *VideoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	if m.SearchVideosByKeywordFn != nil {
		return m.SearchVideosByKeywordFn(ctx, keyword, pageNum, pageSize)
	}
	return nil, 0, nil
}

func (m *VideoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
	if m.GetVideosByIDsFn != nil {
		return m.GetVideosByIDsFn(ctx, videoIDs)
	}
	return nil, nil
}

func (m *VideoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	if m.GetVideosByAuthorIDFn != nil {
		return m.GetVideosByAuthorIDFn(ctx, authorID, pageNum, pageSize)
	}
	return nil, 0, nil
}
