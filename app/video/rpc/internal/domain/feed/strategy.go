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
