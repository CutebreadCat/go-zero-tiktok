package mock

import "context"

type VideoLikerRepo struct {
	LikeVideoFn               func(ctx context.Context, userID, videoID string) error
	CancelLikeVideoFn         func(ctx context.Context, userID, videoID string) error
	GetLikedVideoIDsByUserIDFn func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error)
}

func (m *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	if m.LikeVideoFn != nil {
		return m.LikeVideoFn(ctx, userID, videoID)
	}
	return nil
}

func (m *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if m.CancelLikeVideoFn != nil {
		return m.CancelLikeVideoFn(ctx, userID, videoID)
	}
	return nil
}

func (m *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	if m.GetLikedVideoIDsByUserIDFn != nil {
		return m.GetLikedVideoIDsByUserIDFn(ctx, userID, pageNumber, pageSize)
	}
	return nil, 0, nil
}
