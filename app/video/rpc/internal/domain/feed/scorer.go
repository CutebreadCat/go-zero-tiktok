package feed

import (
	"math"
	"time"

	"go_zero-tiktok/pkg/contract"
)

// ScorerWeights 规则打分权重配置。
type ScorerWeights struct {
	HotWeight     float64 // 热度分权重
	RecencyWeight float64 // 时效性权重
	FollowWeight  float64 // 关注加权
	QoSWeight     float64 // QoS 权重
}

// DefaultScorerWeights 返回默认权重。
func DefaultScorerWeights() ScorerWeights {
	return ScorerWeights{
		HotWeight:     0.50,
		RecencyWeight: 0.20,
		FollowWeight:  0.20,
		QoSWeight:     0.10,
	}
}

// Scorer 规则打分器，用于推荐流粗排。
type Scorer struct {
	weights ScorerWeights
}

// NewScorer 创建规则打分器。
func NewScorer(weights ScorerWeights) *Scorer {
	if weights.HotWeight+weights.RecencyWeight+weights.FollowWeight+weights.QoSWeight == 0 {
		weights = DefaultScorerWeights()
	}
	return &Scorer{weights: weights}
}

// Score 计算单个视频的推荐分。
// 返回 int64 分数，越大越靠前。
func (s *Scorer) Score(video types.VideoBaseinfo, popular types.VideoPopular, followed bool, now time.Time) int64 {
	hotScore := float64(popular.HotScore)
	recencyScore := s.recencyScore(video.CreatedAt, now)

	score := hotScore*s.weights.HotWeight +
		recencyScore*s.weights.RecencyWeight

	// 关注作者加成
	if followed {
		score += 10000 * s.weights.FollowWeight
	}

	// QoS 加成/降权：以 1.0 为中性，高于 1.0 加分，低于 1.0 降分
	qosBoost := s.qosBoost(popular.VideoQoSMetrics)
	score += (qosBoost - 1.0) * 10000 * s.weights.QoSWeight

	return int64(math.Round(score))
}

// recencyScore 计算时效分，越新越高，范围 0~10000。
func (s *Scorer) recencyScore(createdAt string, now time.Time) float64 {
	publishedAt, err := lastPublishedAtMs(createdAt)
	if err != nil || publishedAt <= 0 {
		return 0
	}

	hours := float64(now.UnixMilli()-publishedAt) / (1000 * 60 * 60)
	if hours < 0 {
		hours = 0
	}
	// 1 / (hours + 2) * 10000，保证新视频分高，老视频逐渐衰减
	score := 10000.0 / (hours + 2.0)
	if score > 10000 {
		score = 10000
	}
	return score
}

// followBoost 关注作者加权系数。
func (s *Scorer) followBoost(followed bool) float64 {
	if followed {
		return 1.5
	}
	return 1.0
}

// qosBoost 根据 QoS 指标计算加权系数，范围 0~2。
// 完播率高加分，卡顿率高降分。
func (s *Scorer) qosBoost(metrics types.VideoQoSMetrics) float64 {
	// 无 QoS 数据时给中性分
	if metrics.ReportCount == 0 {
		return 1.0
	}

	// completion_rate 是万分比，转成 0~1
	completion := float64(metrics.CompletionRate) / 10000.0
	stall := float64(metrics.StallRate) / 10000.0
	errors := float64(metrics.ErrorRate) / 10000.0

	boost := 1.0 + completion*0.5 - stall*0.5 - errors*0.5
	if boost < 0 {
		boost = 0
	}
	if boost > 2 {
		boost = 2
	}
	return boost
}
