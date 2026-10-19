package domain

import (
	"context"
	"io"
	"time"

	"go_zero-tiktok/pkg/contract"
)

type IVideoRepo interface {
	// CreateVideoFromParams 创建视频基础信息，返回最终生效的 video_id（幂等键命中时复用已有记录）。
	CreateVideoFromParams(ctx context.Context, videoID, authorID int64, videoURL, coverURL, title, description, idempotencyKey string) (int64, error)
	// GetVideoByAuthorAndIdempotencyKey 按 (author_id, idempotency_key) 查询已发布视频，未命中 found=false。
	GetVideoByAuthorAndIdempotencyKey(ctx context.Context, authorID int64, idempotencyKey string) (videoID int64, publishAt time.Time, found bool, err error)
	SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error)
	GetVideosByAuthorID(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	// GetVideosByCursor 复合游标分页兜底：按 (created_at, video_id) < (publishedAt, videoID) 倒序取 limit 条。
	// publishedAt=0 且 videoID=0 表示首页。
	GetVideosByCursor(ctx context.Context, publishedAt, videoID int64, limit int32) ([]types.VideoBaseinfo, error)
	// GetVideoPublishAt 获取视频发布时间，优先读缓存，未命中回源 MySQL。
	GetVideoPublishAt(ctx context.Context, videoID int64) (time.Time, error)
}

type IPopularRepo interface {
	CreatePopularVideo(ctx context.Context, videoID int64) error
	IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error
	UpdateVideoLikeCount(ctx context.Context, videoID int64, delta int64) error
	UpdateVideoFavoriteCount(ctx context.Context, videoID int64, delta int64) error
	GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error)
	// GetPopularVideosByIDs 批量查询视频热度统计（含 visit_count）。
	GetPopularVideosByIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoPopular, error)
	// SetLikeCount 直接设置视频 like_count（供 syncer 以 Redis 为基准对齐 MySQL）。
	SetLikeCount(ctx context.Context, videoID int64, count int64) error
	// GetLikeCounts 批量查询视频 like_count，用于 Redis 未命中时回源。
	GetLikeCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error)
	// GetPopularVideosByCursor 复合游标分页：按 (visit_count, video_id) < (score, videoID) 倒序取 limit 条。
	// score=0 且 videoID=0 表示首页。
	GetPopularVideosByCursor(ctx context.Context, score, videoID int64, limit int32) ([]types.VideoPopular, error)
	// ListVideoStatsForRebuild 按候选窗口分批读取统计快照。
	// since 为零值表示全量模式（按 video_id 游标）；否则按 (created_at, video_id) 复合游标。
	ListVideoStatsForRebuild(ctx context.Context, since, cursorAt time.Time, cursorID int64, limit int32) ([]types.VideoStatSnapshot, error)
	// CountVideoStatsForRebuild 统计候选窗口内的视频数，供 dry-run 预检。
	CountVideoStatsForRebuild(ctx context.Context, since time.Time) (int64, error)
	// ListVideoStatsByIDs 按视频 ID 批量读取统计快照，供榜单分数保鲜重算使用。
	// 缺少统计行或已软删的视频不会返回。
	ListVideoStatsByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoStatSnapshot, error)
}

type IVideoInteractionRepo interface {
	LikeVideo(ctx context.Context, userID, videoID int64) error
	CancelLikeVideo(ctx context.Context, userID, videoID int64) error
	GetLikedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error)
	FavoriteVideo(ctx context.Context, userID, videoID int64) error
	CancelFavoriteVideo(ctx context.Context, userID, videoID int64) error
	GetFavoritedVideoIDsByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]int64, int64, error)
	// 以下供后台 syncer 批量同步 Redis 关系到 MySQL 使用。
	GetLikeUserIDsByVideoID(ctx context.Context, videoID int64) ([]int64, error)
	BatchAddLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error
	BatchRemoveLikeInteractions(ctx context.Context, videoID int64, userIDs []int64) error
	// ApplyLikeEvent 事务内应用点赞/取消点赞事件（Kafka 消费者落库用）。
	// action 取值与 interaction.LikeAction 一致："like" / "cancel"。
	ApplyLikeEvent(ctx context.Context, action string, userID, videoID int64) error
}

type IPlaybackQoSRepo interface {
	CreateReport(ctx context.Context, report *types.PlaybackQoSReport) error
	// GetReportsAfterID 按 id 游标读取待聚合的上报记录。
	GetReportsAfterID(ctx context.Context, lastID int64, limit int32) ([]*types.PlaybackQoSReport, error)
	// GetReportsByVideoIDs 批量读取指定视频的全部上报记录（用于重算指标）。
	GetReportsByVideoIDs(ctx context.Context, videoIDs []int64) ([]*types.PlaybackQoSReport, error)
}

type IVideoQoSRepo interface {
	// UpdateQoSAggregates 更新视频 QoS 聚合指标（不存在则创建）。
	UpdateQoSAggregates(ctx context.Context, videoID int64, metrics types.VideoQoSMetrics) error
	// GetQoSMetricsByVideoIDs 批量查询视频 QoS 聚合指标。
	GetQoSMetricsByVideoIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoQoSMetrics, error)
}

type StorageProvider interface {
	UploadFile(reader io.Reader, objectKey string) (string, error)
}

// ISeenRepo 用户曝光记录访问接口：记录用户已刷到的视频，用于推荐去重。
// 使用 Redis ZSet 存储，member=video_id，score=曝光时间戳，支持按时间 TTL 淘汰和容量控制。
type ISeenRepo interface {
	// IsSeen 判断指定视频是否已被用户曝光。
	IsSeen(ctx context.Context, userID, videoID int64) (bool, error)
	// MarkSeen 批量标记视频为用户已曝光。
	MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error
	// Cleanup 清理过期和超出容量限制的曝光记录。
	Cleanup(ctx context.Context, userID int64, ttl time.Duration, maxSize int) error
}

// IFeedRepo Feed 索引访问接口：全站候选池（feed:global）+ 关注流收件箱（feed:inbox:{uid}）。
// 索引只存"有序 video_id 索引"，视频详情以 MySQL 为准（索引+水合模式）。
// FeedIndex.Score 为发布时间戳(UnixMilli)，用于跨流按时间倒序合并、去重。
type IFeedRepo interface {
	// AddToGlobalPool 发布成功后写入候选池，并裁剪窗口外成员。
	AddToGlobalPool(ctx context.Context, videoID int64, publishAt time.Time) error
	// GetGlobalPool 取候选池 (lastTimeMs, +inf] 范围内按 score 倒序的索引。
	GetGlobalPool(ctx context.Context, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
	// RemoveFromGlobalPool 从候选池移除视频（下架/删除时调用）。
	RemoveFromGlobalPool(ctx context.Context, videoID int64) error
	// PoolLen 返回候选池当前成员数。
	PoolLen(ctx context.Context) (int64, error)
	// AddToUserInbox 关注流扇出：将视频写入单个用户的收件箱 feed:inbox:{uid}。
	AddToUserInbox(ctx context.Context, uid, videoID int64, publishAt time.Time) error
	// FanoutInbox 批量扇出：将视频一次性写入多个用户的收件箱（pipeline 合并 RTT）。
	FanoutInbox(ctx context.Context, videoID int64, userIDs []int64, publishAt time.Time) error
	// GetUserInbox 取用户收件箱 (lastTimeMs, +inf] 范围内按 score 倒序的索引。
	GetUserInbox(ctx context.Context, uid, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
	// RefreshHotScore 覆盖更新视频热度分。
	RefreshHotScore(ctx context.Context, videoID int64, score int64) error
	// GetHotVideosByCursor 从 hot:videos 取热度分 <= cursorScore 的索引，按热度倒序。
	// 返回数量会略多于 limit，供调用方按 video_id 做同分精断。
	GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error)
	// RemoveHotVideos 从热门榜单移除详情已不存在的成员，避免榜单出现空洞。
	RemoveHotVideos(ctx context.Context, videoIDs []int64) error
	// GetHotScores 批量读取视频当前热度分，未上榜或读取失败按 0 返回。
	GetHotScores(ctx context.Context, videoIDs []int64) (map[int64]int64, error)
	// TrimHotVideos 只保留热度最高的 keepTopN 个成员，榜单退出由排名决定。
	TrimHotVideos(ctx context.Context, keepTopN int) error

	// 以下供热度榜全量重建使用：先写临时键，构建完成后原子替换正式键。
	ResetHotRebuild(ctx context.Context) error
	AppendHotRebuildScores(ctx context.Context, scores []types.FeedIndex) error
	// PublishHotRebuild 原子替换 hot:videos 并按 keepTopN 裁剪，返回最终成员数。
	PublishHotRebuild(ctx context.Context, keepTopN int) (int64, error)
	// TryLockHotRebuild 抢占重建锁并返回本次执行的 token；ttl 为锁自动过期时间。
	TryLockHotRebuild(ctx context.Context, ttl time.Duration) (token string, ok bool, err error)
	// UnlockHotRebuild 仅在 token 匹配时释放锁，避免误删其他执行者的锁。
	UnlockHotRebuild(ctx context.Context, token string) error
	// HotRebuildRunning 探测是否有重建任务正在执行，供分数保鲜任务避让。
	HotRebuildRunning(ctx context.Context) (bool, error)
	// RefreshHotScores 批量覆盖正式榜单分数，仅更新已存在成员，不新增成员。
	RefreshHotScores(ctx context.Context, scores []types.FeedIndex) error
}
