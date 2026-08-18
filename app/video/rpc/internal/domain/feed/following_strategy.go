package feed

import (
	"context"

	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

// FollowingStrategy 关注流策略：只返回当前用户关注的人发布的视频。
// 数据源为 feed:inbox:{uid}，不走全站候选池，也不走 MySQL 兜底。
type FollowingStrategy struct {
	videoRepo   domainVideoRepo
	popularRepo domainPopularRepo
	feedRepo    domainFeedRepo
}

// NewFollowingStrategy 创建关注流策略。
func NewFollowingStrategy(videoRepo domainVideoRepo, popularRepo domainPopularRepo, feedRepo domainFeedRepo) *FollowingStrategy {
	return &FollowingStrategy{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		feedRepo:    feedRepo,
	}
}

// Name 返回策略名。
func (s *FollowingStrategy) Name() string {
	return "following"
}

// GetFeed 读取关注流 Feed 一页。
func (s *FollowingStrategy) GetFeed(ctx context.Context, viewerID int64, cursor string, limit int32) (*Result, error) {
	if viewerID <= 0 {
		return nil, xerr.NewInvalidParam("following feed 需要登录态")
	}
	if s.feedRepo == nil {
		return nil, xerr.NewServerBusy()
	}
	if limit <= 0 || limit > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	parsedCursor, err := DecodeTimelineCursor(cursor)
	if err != nil {
		return nil, err
	}

	var lastTimeMs int64
	if parsedCursor != nil {
		lastTimeMs = parsedCursor.PublishedAt
	}

	indexes, err := s.feedRepo.GetUserInbox(ctx, viewerID, lastTimeMs, int(limit)+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(indexes) > int(limit)
	if hasMore {
		indexes = indexes[:limit]
	}

	if len(indexes) == 0 {
		return &Result{
			Videos:   nil,
			Populars: nil,
			HasMore:  false,
			Total:    0,
		}, nil
	}

	videoIDs := make([]int64, 0, len(indexes))
	for _, idx := range indexes {
		videoIDs = append(videoIDs, idx.VideoID)
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	populars := batchGetPopulars(ctx, videos, s.popularRepo)

	var nextCursor string
	if len(videos) > 0 {
		last := videos[len(videos)-1]
		lastMs, err := lastPublishedAtMs(last.CreatedAt)
		if err != nil {
			logx.Errorf("parse last video created_at failed: %v", err)
		} else {
			nextCursor = EncodeTimelineCursor(&TimelineCursor{
				PublishedAt: lastMs,
				VideoID:     last.VideoID,
			})
		}
	}

	return &Result{
		Videos:     videos,
		Populars:   populars,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      int64(len(videos)),
	}, nil
}
