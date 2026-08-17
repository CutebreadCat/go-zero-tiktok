package domain

import (
	"context"
	"io"
	"time"

	"go_zero-tiktok/pkg/contract"
)

type IVideoRepo interface {
	CreateVideoFromParams(ctx context.Context, videoID, authorID int64, videoURL, coverURL, title, description string) error
	SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error)
	GetVideosByAuthorID(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
	GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
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

type StorageProvider interface {
	UploadFile(reader io.Reader, objectKey string) (string, error)
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
}
