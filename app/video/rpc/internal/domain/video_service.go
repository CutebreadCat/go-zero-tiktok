package domain

import (
	"context"
	"strings"
	"time"

	"go_zero-tiktok/pkg/contract"
	myutils "go_zero-tiktok/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// VideoService 视频领域服务，聚焦视频发布、Feed、搜索、热度排序等核心能力。
// 点赞/收藏等互动能力已拆分到 interaction 子领域。
type VideoService struct {
	videoRepo   IVideoRepo
	popularRepo IPopularRepo
	storage     StorageProvider
	feedRepo    IFeedRepo
}

func NewVideoService(videoRepo IVideoRepo, popularRepo IPopularRepo, storage StorageProvider, feedRepo IFeedRepo) *VideoService {
	return &VideoService{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		storage:     storage,
		feedRepo:    feedRepo,
	}
}

// PublishVideo 创建视频并初始化 Popular 记录，发布成功后写入 Feed 候选池。
// 候选池写入失败不阻断发布（DB 是主数据源，Redis 是加速层）。
func (s *VideoService) PublishVideo(ctx context.Context, videoID, authorID int64, videoObjectKey, coverObjectKey, title, description string) error {
	if err := s.videoRepo.CreateVideoFromParams(ctx, videoID, authorID, videoObjectKey, coverObjectKey, title, description); err != nil {
		return err
	}
	if err := s.popularRepo.CreatePopularVideo(ctx, videoID); err != nil {
		return err
	}

	// 写入候选池 feed:global（失败仅记日志，不影响发布主流程）
	if s.feedRepo != nil {
		if err := s.feedRepo.AddToGlobalPool(ctx, videoID, time.Now()); err != nil {
			logx.Errorf("add video %d to feed global pool failed: %v", videoID, err)
		}
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

func (s *VideoService) GetFeedVideos(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	// 优先走候选池（Redis ZSet 索引 + MySQL 水合）
	if s.feedRepo != nil {
		videos, total, err := s.getFeedFromGlobalPool(ctx, lastTime, pageSize)
		if err != nil {
			logx.Errorf("get feed from global pool failed, fallback to db: %v", err)
		} else if len(videos) > 0 {
			s.recordVisits(videos)
			return videos, total, nil
		}
	}

	// 兜底：MySQL 直查（现有逻辑）
	videos, total, err := s.videoRepo.GetVideoByLastTime(ctx, lastTime, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(videos)
	return videos, total, nil
}

// getFeedFromGlobalPool 从候选池取视频：先 ZREVRANGEBYSCORE 拿有序 id，再 GetVideosByIDs 水合详情。
// 候选池只保留最近 7 天，因此窗口外视频由 MySQL 兜底补齐，保证不空页。
func (s *VideoService) getFeedFromGlobalPool(ctx context.Context, lastTime string, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	// lastTime 字符串转毫秒时间戳，作为 ZSet 游标
	var lastTimeMs int64
	if lt := strings.TrimSpace(lastTime); lt != "" {
		t, err := myutils.StrToTime(lt, "")
		if err != nil {
			return nil, 0, err
		}
		lastTimeMs = t.UnixMilli()
	}

	// 多取一条判断是否还有更多（用于 total 边界）
	ids, err := s.feedRepo.GetGlobalPoolIDs(ctx, lastTimeMs, int(pageSize)+1)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(ids))
	hasMore := len(ids) > int(pageSize)
	if hasMore {
		ids = ids[:pageSize]
	}
	if len(ids) == 0 {
		// 候选池为空（冷启动/无视频），交由调用方走 MySQL 兜底
		return nil, 0, nil
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	if hasMore {
		total = int64(len(videos))
	}
	return videos, total, nil
}
