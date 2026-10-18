package feed

import (
	"context"

	"go_zero-tiktok/pkg/contract"
)

// Strategy 定义 Feed 场景策略接口。
type Strategy interface {
	// Name 返回策略名，用于注册与日志。
	Name() string
	// GetFeed 根据游标读取一页 Feed，返回视频列表、热度统计、下一页游标与是否还有更多。
	GetFeed(ctx context.Context, viewerID int64, cursor string, limit int32) (*Result, error)
}

// Result 是 Feed 策略的统一返回结构。
type Result struct {
	Videos     []types.VideoBaseinfo
	Populars   []types.VideoPopular
	NextCursor string
	HasMore    bool
	Total      int64
}

// domainVideoRepo 是各 Strategy 所需的视频仓储能力子集。
type domainVideoRepo interface {
	GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error)
	// GetVideosByCursor 复合游标分页兜底：按 (created_at, video_id) < (publishedAt, videoID) 倒序取 limit 条。
	GetVideosByCursor(ctx context.Context, publishedAt, videoID int64, limit int32) ([]types.VideoBaseinfo, error)
}

// domainPopularRepo 是各 Strategy 所需的热度仓储能力子集。
type domainPopularRepo interface {
	GetPopularVideosByIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoPopular, error)
	// GetPopularVideosByCursor 复合游标分页：按 (visit_count, video_id) < (score, videoID) 倒序取 limit 条。
	GetPopularVideosByCursor(ctx context.Context, score, videoID int64, limit int32) ([]types.VideoPopular, error)
}

// domainFeedRepo 是各 Strategy 所需的 Feed 索引仓储能力子集。
type domainFeedRepo interface {
	// GetGlobalPool 取全站候选池索引。
	GetGlobalPool(ctx context.Context, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
	// GetUserInbox 取用户关注收件箱索引。
	GetUserInbox(ctx context.Context, uid, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
	// GetHotVideosByCursor 从 hot:videos 取热度索引。
	GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error)
	// GetHotScores 批量读取视频热度分，用于推荐粗排的热度加权。
	GetHotScores(ctx context.Context, videoIDs []int64) (map[int64]int64, error)
}

// domainHotIndexRepo 是热门排行榜所需的索引能力子集。
// 除读取排名外，还需要在视频详情缺失时移除失效成员（榜单自愈）。
type domainHotIndexRepo interface {
	GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error)
	RemoveHotVideos(ctx context.Context, videoIDs []int64) error
}

// domainQoSRepo 是各 Strategy 所需的 QoS 指标仓储能力子集。
type domainQoSRepo interface {
	// GetQoSMetricsByVideoIDs 批量查询视频 QoS 聚合指标。
	GetQoSMetricsByVideoIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoQoSMetrics, error)
}

// domainSeenRepo 是各 Strategy 所需的曝光记录仓储能力子集。
type domainSeenRepo interface {
	// IsSeen 判断指定视频是否已被用户曝光。
	IsSeen(ctx context.Context, userID, videoID int64) (bool, error)
	// MarkSeen 批量标记视频为用户已曝光。
	MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error
}
