package repository

import (
	"context"

	videobasetable "go_zero-tiktok/internal/dal/tables/video_baseinfo"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type VideoBaseinfoRepo struct {
	db *gorm.DB
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{db: db}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *types.VideoBaseinfo) error {
	return videobasetable.CreateVideo(ctx, r.db, video)
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return videobasetable.SearchVideosByKeyword(ctx, r.db, keyword, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
	return videobasetable.GetVideosByIDs(ctx, r.db, videoIDs)
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return videobasetable.GetVideosByAuthorID(ctx, r.db, authorID, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByVisitCount(ctx context.Context, pageNum, pageSize int32, videoIDs []string) ([]types.VideoBaseinfo, int64, error) {
	return videobasetable.GetVideosByVisitCount(ctx, r.db, pageNum, pageSize, videoIDs)
}
