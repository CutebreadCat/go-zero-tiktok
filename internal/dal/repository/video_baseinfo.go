package repository

import (
	"context"

	videobasetable "go_zero-tiktok/internal/dal/tables/video_baseinfo"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type VideoBaseinfoRepo struct {
	db *gorm.DB
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{db: db}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *videobasetable.VideoBaseinfo) error {
	if err := videobasetable.CreateVideo(ctx, r.db, video); err != nil {
		return pkgerrors.WithMessage(err, "VideoBaseinfoRepo.CreateVideo")
	}
	return nil
}

func (r *VideoBaseinfoRepo) CreateVideoFromParams(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error {
	video := &videobasetable.VideoBaseinfo{
		VideoID:     videoID,
		AuthorID:    authorID,
		VideoURL:    videoURL,
		CoverURL:    coverURL,
		Title:       title,
		Description: description,
	}
	if err := videobasetable.CreateVideo(ctx, r.db, video); err != nil {
		return pkgerrors.WithMessage(err, "VideoBaseinfoRepo.CreateVideoFromParams")
	}
	return nil
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]videobasetable.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.SearchVideosByKeyword(ctx, r.db, keyword, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.SearchVideosByKeyword")
	}
	return videos, total, nil
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]videobasetable.VideoBaseinfo, error) {
	videos, err := videobasetable.GetVideosByIDs(ctx, r.db, videoIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByIDs")
	}
	return videos, nil
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]videobasetable.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.GetVideosByAuthorID(ctx, r.db, authorID, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByAuthorID")
	}
	return videos, total, nil
}

func (r *VideoBaseinfoRepo) GetVideosByVisitCount(ctx context.Context, pageNum, pageSize int32, videoIDs []string) ([]videobasetable.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.GetVideosByVisitCount(ctx, r.db, pageNum, pageSize, videoIDs)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByVisitCount")
	}
	return videos, total, nil
}

func (r *VideoBaseinfoRepo) VideoToResponse(video *videobasetable.VideoBaseinfo) types.VideoBaseinfo {
	return types.VideoBaseinfo{
		VideoID:     video.VideoID,
		AuthorID:    video.AuthorID,
		VideoURL:    video.VideoURL,
		CoverURL:    video.CoverURL,
		Title:       video.Title,
		Description: video.Description,
		CreatedAt:   myutils.TimeToStr(video.CreatedAt, ""),
		UpdatedAt:   myutils.TimeToStr(video.UpdatedAt, ""),
		DeletedAt:   myutils.NullTimeToStr(myutils.TimeToNullTime(video.DeletedAt), ""),
	}
}

func (r *VideoBaseinfoRepo) VideosToResponse(videos []videobasetable.VideoBaseinfo) []types.VideoBaseinfo {
	result := make([]types.VideoBaseinfo, 0, len(videos))
	for _, v := range videos {
		result = append(result, r.VideoToResponse(&v))
	}
	return result
}
