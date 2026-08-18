package domain

import (
	"context"
	"time"

	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/pkg/contract"

	"github.com/zeromicro/go-zero/core/logx"
)

// VideoService 视频领域服务，聚焦视频发布、Feed、搜索、热度排序等核心能力。
// 点赞/收藏等互动能力已拆分到 interaction 子领域。
type VideoService struct {
	videoRepo       IVideoRepo
	popularRepo     IPopularRepo
	storage         StorageProvider
	feedRepo        IFeedRepo
	strategyFactory *feedpkg.StrategyFactory
}

func NewVideoService(videoRepo IVideoRepo, popularRepo IPopularRepo, storage StorageProvider, feedRepo IFeedRepo) *VideoService {
	return &VideoService{
		videoRepo:       videoRepo,
		popularRepo:     popularRepo,
		storage:         storage,
		feedRepo:        feedRepo,
		strategyFactory: feedpkg.NewStrategyFactory(feedpkg.NewTimelineStrategy(videoRepo, popularRepo, feedRepo)),
	}
}

// PublishVideo 创建视频并初始化 Popular 记录，发布成功后写入 Feed 候选池。
// 候选池写入失败不阻断发布（DB 是主数据源，Redis 是加速层）。
// 返回发布时刻 publishAt，供 gateway 编排扇出时保证 feed:global 与 feed:inbox 的 score 一致。
func (s *VideoService) PublishVideo(ctx context.Context, videoID, authorID int64, videoObjectKey, coverObjectKey, title, description string) (time.Time, error) {
	publishAt := time.Now()
	if err := s.videoRepo.CreateVideoFromParams(ctx, videoID, authorID, videoObjectKey, coverObjectKey, title, description); err != nil {
		return time.Time{}, err
	}
	if err := s.popularRepo.CreatePopularVideo(ctx, videoID); err != nil {
		return time.Time{}, err
	}

	// 写入候选池 feed:global（失败仅记日志，不影响发布主流程）
	if s.feedRepo != nil {
		if err := s.feedRepo.AddToGlobalPool(ctx, videoID, publishAt); err != nil {
			logx.Errorf("add video %d to feed global pool failed: %v", videoID, err)
		}
	}
	return publishAt, nil
}

// FanoutToUsers 关注流扇出：将视频写入每个粉丝的收件箱 feed:inbox:{uid}。
// 扇出属于"尽力而为"——失败仅记日志返回错误，由调用方决定是否阻断发布（设计上不应阻断）。
// 粉丝列表由 gateway 编排层从 communication.rpc 获取，本方法只写 video.rpc 自己的 Redis 索引。
func (s *VideoService) FanoutToUsers(ctx context.Context, videoID int64, userIDs []int64, publishAt time.Time) error {
	if s.feedRepo == nil || len(userIDs) == 0 {
		return nil
	}
	if err := s.feedRepo.FanoutInbox(ctx, videoID, userIDs, publishAt); err != nil {
		logx.Errorf("fanout video %d to %d users failed: %v", videoID, len(userIDs), err)
		return err
	}
	return nil
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

// GetFeedVideos 获取 Feed 流。viewerID 为当前浏览用户（<=0 表示无登录态，仅看全站候选池）。
// 根据 scene 路由到对应策略，默认 scene 为 timeline。
// 返回视频基础信息、热度统计、下一页游标、是否还有更多。
func (s *VideoService) GetFeedVideos(ctx context.Context, viewerID int64, scene, cursor string, limit int32) (*feedpkg.Result, error) {
	strategy, err := s.strategyFactory.Get(scene)
	if err != nil {
		return nil, err
	}

	result, err := strategy.GetFeed(ctx, viewerID, cursor, limit)
	if err != nil {
		return nil, err
	}

	s.recordVisits(result.Videos)
	return result, nil
}
