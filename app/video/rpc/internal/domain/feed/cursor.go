package feed

import (
	"strings"

	"go_zero-tiktok/pkg/cursor"
	"go_zero-tiktok/pkg/xerr"
)

// TimelineCursor 是时间线/关注流场景的复合游标。
// 使用 published_at + video_id 组合，避免同一毫秒下多条数据翻页冲突。
type TimelineCursor struct {
	PublishedAt int64 `json:"published_at"`
	VideoID     int64 `json:"video_id"`
}

// EncodeTimelineCursor 把时间线游标编码为字符串。
// 游标无效时返回空字符串。
func EncodeTimelineCursor(c *TimelineCursor) string {
	if c == nil || c.VideoID <= 0 {
		return ""
	}

	encoded, err := cursor.Encode(c)
	if err != nil {
		return ""
	}
	return encoded
}

// DecodeTimelineCursor 解码时间线游标。
// 空游标表示首页，返回 nil。
func DecodeTimelineCursor(raw string) (*TimelineCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var c TimelineCursor
	if err := cursor.Decode(raw, &c); err != nil {
		return nil, xerr.NewInvalidParam("invalid timeline cursor")
	}
	if c.VideoID <= 0 {
		return nil, xerr.NewInvalidParam("invalid timeline cursor")
	}
	return &c, nil
}

// HotCursor 是热门场景的复合游标。
// 使用 score + video_id 组合，score 为热度分，video_id 用于同分精断。
type HotCursor struct {
	Score   int64 `json:"score"`
	VideoID int64 `json:"video_id"`
}

// EncodeHotCursor 把热门游标编码为字符串。
func EncodeHotCursor(c *HotCursor) string {
	if c == nil || c.VideoID <= 0 {
		return ""
	}
	encoded, err := cursor.Encode(c)
	if err != nil {
		return ""
	}
	return encoded
}

// DecodeHotCursor 解码热门游标。
// 空游标表示首页，返回 nil。
func DecodeHotCursor(raw string) (*HotCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var c HotCursor
	if err := cursor.Decode(raw, &c); err != nil {
		return nil, xerr.NewInvalidParam("invalid hot cursor")
	}
	if c.VideoID <= 0 {
		return nil, xerr.NewInvalidParam("invalid hot cursor")
	}
	return &c, nil
}

// RecommendCursor 是推荐场景的复合游标。
// 使用 recommend_score + video_id 组合，score 为推荐粗排分，video_id 用于同分精断。
type RecommendCursor struct {
	Score   int64 `json:"score"`
	VideoID int64 `json:"video_id"`
}

// EncodeRecommendCursor 把推荐游标编码为字符串。
// 游标无效时返回空字符串。
func EncodeRecommendCursor(c *RecommendCursor) string {
	if c == nil || c.VideoID <= 0 {
		return ""
	}
	encoded, err := cursor.Encode(c)
	if err != nil {
		return ""
	}
	return encoded
}

// DecodeRecommendCursor 解码推荐游标。
// 空游标表示首页，返回 nil。
func DecodeRecommendCursor(raw string) (*RecommendCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var c RecommendCursor
	if err := cursor.Decode(raw, &c); err != nil {
		return nil, xerr.NewInvalidParam("invalid recommend cursor")
	}
	if c.VideoID <= 0 {
		return nil, xerr.NewInvalidParam("invalid recommend cursor")
	}
	return &c, nil
}
