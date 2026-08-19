package feed

import (
	"math"
	"time"

	"go_zero-tiktok/pkg/contract"
)

// HotScoreCalculator 热度分计算器，使用简化版 Hacker News 公式。
// 热度分 = (base_score + visit + w_like*like + w_comment*comment + w_favorite*favorite)
//          / ((age_hours + 2) ^ gravity)
type HotScoreCalculator struct {
	BaseScore      float64
	LikeWeight     float64
	CommentWeight  float64
	FavoriteWeight float64
	Gravity        float64
}

// Compute 计算视频热度分，返回放大 1000 倍后的 int64。
// 放大是为了在 Redis ZSet 中使用整数 score，同时保持 HotCursor 仍为 int64。
func (c *HotScoreCalculator) Compute(popular types.VideoPopular, publishAt time.Time) int64 {
	engagement := c.BaseScore +
		float64(popular.VisitCount) +
		c.LikeWeight*float64(popular.LikeCount) +
		c.CommentWeight*float64(popular.CommentCount) +
		c.FavoriteWeight*float64(popular.FavoriteCount)

	hours := time.Since(publishAt).Hours()
	if hours < 0 {
		hours = 0
	}

	score := engagement / math.Pow(hours+2, c.Gravity)
	return int64(score * 1000)
}
