package repository

import (
	"context"

	videobasetable "go_zero-tiktok/internal/dal/tables/video_baseinfo"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"gorm.io/gorm"
)

type VideoBaseinfoRepo struct {
	db *gorm.DB
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{db: db}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *videobasetable.VideoBaseinfo) error {
	return videobasetable.CreateVideo(ctx, r.db, video)
}

// CreateVideoFromParams 通过参数创建视频，logic层不需要知道数据库模型
func (r *VideoBaseinfoRepo) CreateVideoFromParams(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error {
	video := &videobasetable.VideoBaseinfo{
		VideoID:     videoID,
		AuthorID:    authorID,
		VideoURL:    videoURL,
		CoverURL:    coverURL,
		Title:       title,
		Description: description,
	}
	return videobasetable.CreateVideo(ctx, r.db, video)
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]videobasetable.VideoBaseinfo, int64, error) {
	return videobasetable.SearchVideosByKeyword(ctx, r.db, keyword, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]videobasetable.VideoBaseinfo, error) {
	return videobasetable.GetVideosByIDs(ctx, r.db, videoIDs)
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]videobasetable.VideoBaseinfo, int64, error) {
	return videobasetable.GetVideosByAuthorID(ctx, r.db, authorID, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByVisitCount(ctx context.Context, pageNum, pageSize int32, videoIDs []string) ([]videobasetable.VideoBaseinfo, int64, error) {
	return videobasetable.GetVideosByVisitCount(ctx, r.db, pageNum, pageSize, videoIDs)
}

// VideoToResponse 将数据库模型转换为API响应类型
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

// VideosToResponse 将数据库模型切片转换为API响应类型切片
func (r *VideoBaseinfoRepo) VideosToResponse(videos []videobasetable.VideoBaseinfo) []types.VideoBaseinfo {
	result := make([]types.VideoBaseinfo, 0, len(videos))
	for _, v := range videos {
		result = append(result, r.VideoToResponse(&v))
	}
	return result
}

