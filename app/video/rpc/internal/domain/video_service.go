package domain

import (
	"context"
	"time"

	"go_zero-tiktok/pkg/contract"

	"github.com/zeromicro/go-zero/core/logx"
)

// VideoService 视频领域服务，聚焦视频发布、Feed、搜索、热度排序等核心能力。
// 点赞/收藏等互动能力已拆分到 interaction 子领域。
type VideoService struct {
	videoRepo   IVideoRepo
	popularRepo IPopularRepo
	storage     StorageProvider
}

func NewVideoService(videoRepo IVideoRepo, popularRepo IPopularRepo, storage StorageProvider) *VideoService {
	return &VideoService{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		storage:     storage,
	}
}

// PublishVideo 创建视频并初始化 Popular 记录
func (s *VideoService) PublishVideo(ctx context.Context, videoID, authorID int64, videoObjectKey, coverObjectKey, title, description string) error {
	if err := s.videoRepo.CreateVideoFromParams(ctx, videoID, authorID, videoObjectKey, coverObjectKey, title, description); err != nil {
		return err
	}
	return s.popularRepo.CreatePopularVideo(ctx, videoID)
}

// IncreaseVideoVisitCount 同步增加视频访问量（供 interaction 服务跨 RPC 调用）
func (s *VideoService) IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error {
	return s.popularRepo.IncreaseVideoVisitCount(ctx, videoID, delta)
}

// RecordVisit 异步记录视频访问量
func (s *VideoService) RecordVisit(videoID int64) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in RecordVisit videoID=%d: %v", videoID, r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := s.popularRepo.IncreaseVideoVisitCount(ctx, videoID, 1); err != nil {
			logx.Errorf("increment visit count failed for video %d: %v", videoID, err)
		}
	}()
}

// recordVisits 批量异步记录访问量
func (s *VideoService) recordVisits(videos []types.VideoBaseinfo) {
	for _, video := range videos {
		s.RecordVisit(video.VideoID)
	}
}

// GetVideosByIDs 根据视频 ID 列表批量查询视频信息。
func (s *VideoService) GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error) {
	return s.videoRepo.GetVideosByIDs(ctx, videoIDs)
}

// GetPopularVideos 获取热门视频列表（ID 水合为完整视频信息）
func (s *VideoService) GetPopularVideos(ctx context.Context, pageNum, pageSize int32) ([]types.VideoBaseinfo, []types.VideoPopular, error) {
	videoPopulars, _, err := s.popularRepo.GetPopularVideoIDsByVisitCount(ctx, pageNum, pageSize)
	if err != nil {
		return nil, nil, err
	}

	videoIDs := make([]int64, 0, len(videoPopulars))
	for _, vp := range videoPopulars {
		videoIDs = append(videoIDs, vp.VideoID)
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, nil, err
	}

	return videos, videoPopulars, nil
}

// SearchVideos 搜索视频并记录访问量
func (s *VideoService) SearchVideos(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.SearchVideosByKeyword(ctx, keyword, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(videos)
	return videos, total, nil
}

// GetVideosByAuthor 获取作者视频列表并记录访问量
func (s *VideoService) GetVideosByAuthor(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.GetVideosByAuthorID(ctx, authorID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(videos)
	return videos, total, nil
}

func (s *VideoService) GetFeedVideos(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.GetVideoByLastTime(ctx, lastTime, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(videos)
	return videos, total, nil
}
