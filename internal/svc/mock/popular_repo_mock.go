package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type PopularRepo struct {
	CreatePopularVideoFn               func(ctx context.Context, videoID string) error
	IncreaseVideoVisitCountFn          func(ctx context.Context, videoID string, delta int64) error
	UpdateVideoLikeCountFn             func(ctx context.Context, videoID string, delta int64) error
	GetPopularVideoIDsByVisitCountFn   func(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error)
}

func (m *PopularRepo) CreatePopularVideo(ctx context.Context, videoID string) error {
	if m.CreatePopularVideoFn != nil {
		return m.CreatePopularVideoFn(ctx, videoID)
	}
	return nil
}

func (m *PopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	if m.IncreaseVideoVisitCountFn != nil {
		return m.IncreaseVideoVisitCountFn(ctx, videoID, delta)
	}
	return nil
}

func (m *PopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	if m.UpdateVideoLikeCountFn != nil {
		return m.UpdateVideoLikeCountFn(ctx, videoID, delta)
	}
	return nil
}

func (m *PopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
	if m.GetPopularVideoIDsByVisitCountFn != nil {
		return m.GetPopularVideoIDsByVisitCountFn(ctx, pageNum, pageSize)
	}
	return nil, 0, nil
}
