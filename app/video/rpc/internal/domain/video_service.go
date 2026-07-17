package domain

import (
	"context"
	"fmt"
	"time"

	"go_zero-tiktok/pkg/contract"
)

type VideoService struct {
	videoRepo   IVideoRepo
	popularRepo IPopularRepo
	likerRepo   IVideoLikerRepo
	storage     StorageProvider
}

func NewVideoService(videoRepo IVideoRepo, popularRepo IPopularRepo, likerRepo IVideoLikerRepo, storage StorageProvider) *VideoService {
	return &VideoService{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		likerRepo:   likerRepo,
		storage:     storage,
	}
}

// PublishVideo 创建视频并初始化 Popular 记录
func (s *VideoService) PublishVideo(ctx context.Context, videoID, authorID, videoURL, coverURL, title, description string) error {
	if err := s.videoRepo.CreateVideoFromParams(ctx, videoID, authorID, videoURL, coverURL, title, description); err != nil {
		return err
	}
	return s.popularRepo.CreatePopularVideo(ctx, videoID)
}

// LikeVideo 点赞视频，同步更新 like count
func (s *VideoService) LikeVideo(ctx context.Context, userID, videoID string) error {
	if err := s.likerRepo.LikeVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoLikeCount(ctx, videoID, 1)
}

// CancelLikeVideo 取消点赞，同步更新 like count
func (s *VideoService) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if err := s.likerRepo.CancelLikeVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoLikeCount(ctx, videoID, -1)
}

// RecordVisit 异步记录视频访问量
func (s *VideoService) RecordVisit(videoID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := s.popularRepo.IncreaseVideoVisitCount(ctx, videoID, 1); err != nil {
			fmt.Printf("increment visit count failed for video %s: %v\n", videoID, err)
		}
	}()
}

// GetLikedVideos 获取用户点赞的视频列表（ID 水合为完整视频信息）
func (s *VideoService) GetLikedVideos(ctx context.Context, userID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videoIDs, total, err := s.likerRepo.GetLikedVideoIDsByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// GetPopularVideos 获取热门视频列表（ID 水合为完整视频信息）
func (s *VideoService) GetPopularVideos(ctx context.Context, pageNum, pageSize int32) ([]types.VideoBaseinfo, []types.VideoPopular, error) {
	videoPopulars, _, err := s.popularRepo.GetPopularVideoIDsByVisitCount(ctx, pageNum, pageSize)
	if err != nil {
		return nil, nil, err
	}

	videoIDs := make([]string, 0, len(videoPopulars))
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
	for _, video := range videos {
		s.RecordVisit(video.VideoID)
	}
	return videos, total, nil
}

// GetVideosByAuthor 获取作者视频列表并记录访问量
func (s *VideoService) GetVideosByAuthor(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.GetVideosByAuthorID(ctx, authorID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, video := range videos {
		s.RecordVisit(video.VideoID)
	}
	return videos, total, nil
}

func (s *VideoService) GetFeedVideos(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.GetVideoByLastTime(ctx, lastTime, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, video := range videos {
		s.RecordVisit(video.VideoID)
	}
	return videos, total, nil
}

func (s *VideoService) GetVideosByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return s.GetFeedVideos(ctx, lastTime, pageNum, pageSize)
}
