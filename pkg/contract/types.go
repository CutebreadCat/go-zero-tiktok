// Package types 提供 RPC 服务之间共享的数据契约（DTO）。
//
// 本文件为手工维护，与 api/ 目录下的接口定义独立演化：
//   - HTTP 接口层的请求/响应类型请定义在网关（app/gateway/api/internal/types），由 goctl 生成；
//   - 此处仅存放跨 RPC 服务复用的领域模型，字段以 api/model.api 为基准对齐。
package types

// UserBaseinfo 用户基础信息（user / communication RPC 共享）
type UserBaseinfo struct {
	UserID    int64  `json:"user_id,string"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	PhotoURL  string `json:"photo_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

// UserFollow 用户关注关系（communication RPC 内部使用）
type UserFollow struct {
	FollowerID int64 `json:"follower_id,string"`
	UserID     int64 `json:"user_id,string"`
	Status     int32 `json:"status"`
}

// VideoBaseinfo 视频基础信息（video RPC 使用）
type VideoBaseinfo struct {
	VideoID     int64  `json:"video_id,string"`
	AuthorID    int64  `json:"author_id,string"`
	VideoURL    string `json:"video_url"`
	CoverURL    string `json:"cover_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at"`
}

// FeedIndex Feed 流索引项（feed:global / feed:inbox:{uid} ZSet 的成员 + score）。
// Score 为发布时间戳(UnixMilli)，用于跨流按时间倒序合并。
type FeedIndex struct {
	VideoID int64 `json:"video_id,string"`
	Score   int64 `json:"score"`
}

// VideoPopular 视频热度统计（video RPC 使用）
type VideoPopular struct {
	VideoID       int64 `json:"video_id,string"`
	VisitCount    int64 `json:"visit_count"`
	LikeCount     int64 `json:"like_count"`
	CommentCount  int64 `json:"comment_count"`
	FavoriteCount int64 `json:"favorite_count"`
	HotScore      int64 `json:"hot_score"`
}

// PlaybackQoSReport 播放质量上报领域模型（video RPC 使用）
type PlaybackQoSReport struct {
	UserID         int64
	VideoID        int64
	IdempotencyKey string
	EventType      string
	DurationMs     int64
	PlayedMs       int64
	BufferedMs     int64
	StallCount     int32
	StallTotalMs   int64
	Resolution     string
	BitrateKbps    int32
	Fps            int32
	ErrorCode      int32
	ErrorMsg       string
	NetworkType    string
	DeviceInfo     string
}

// CommentBaseinfo 评论基础信息（interaction RPC 使用）
type CommentBaseinfo struct {
	CommentID       int64  `json:"comment_id,string"`
	UserID          int64  `json:"user_id,string"`
	VideoID         int64  `json:"video_id,string"`
	Content         string `json:"content"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	DeletedAt       string `json:"deleted_at"`
	LikeCount       int32  `json:"like_count"`
	ParentCommentID int64  `json:"parent_comment_id,string"`
}
