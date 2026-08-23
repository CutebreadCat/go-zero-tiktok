package feed

import (
	"testing"
	"time"

	"go_zero-tiktok/pkg/contract"
)

func TestScorer_Score(t *testing.T) {
	scorer := NewScorer(DefaultScorerWeights())
	now := time.Now()

	tests := []struct {
		name     string
		video    types.VideoBaseinfo
		popular  types.VideoPopular
		followed bool
		wantMin  int64
		wantMax  int64
	}{
		{
			name: "高热度视频分高",
			video: types.VideoBaseinfo{
				VideoID:   1,
				AuthorID:  100,
				CreatedAt: now.Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			},
			popular: types.VideoPopular{
				VideoID:  1,
				HotScore: 10000,
			},
			followed: false,
			wantMin:  5000,
			wantMax:  6000,
		},
		{
			name: "新视频时效分高",
			video: types.VideoBaseinfo{
				VideoID:   2,
				AuthorID:  101,
				CreatedAt: now.Add(-10 * time.Minute).Format("2006-01-02 15:04:05"),
			},
			popular: types.VideoPopular{
				VideoID:  2,
				HotScore: 0,
			},
			followed: false,
			wantMin:  500,
			wantMax:  1500,
		},
		{
			name: "关注作者有加成",
			video: types.VideoBaseinfo{
				VideoID:   3,
				AuthorID:  102,
				CreatedAt: now.Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			},
			popular: types.VideoPopular{
				VideoID:  3,
				HotScore: 1000,
			},
			followed: true,
			wantMin:  2500,
			wantMax:  3500,
		},
		{
			name: "QoS 完播率高有加成",
			video: types.VideoBaseinfo{
				VideoID:   4,
				AuthorID:  103,
				CreatedAt: now.Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			},
			popular: types.VideoPopular{
				VideoID: 4,
				VideoQoSMetrics: types.VideoQoSMetrics{
					CompletionRate: 8000, // 80%
					ReportCount:    10,
				},
			},
			followed: false,
			wantMin:  200,
			wantMax:  1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorer.Score(tt.video, tt.popular, tt.followed, now)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Score() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScorer_qosBoost(t *testing.T) {
	scorer := NewScorer(DefaultScorerWeights())

	tests := []struct {
		name    string
		metrics types.VideoQoSMetrics
		wantMin float64
		wantMax float64
	}{
		{
			name:    "无 QoS 数据",
			metrics: types.VideoQoSMetrics{},
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name: "完播率高",
			metrics: types.VideoQoSMetrics{
				CompletionRate: 9000,
				ReportCount:    10,
			},
			wantMin: 1.4,
			wantMax: 1.5,
		},
		{
			name: "卡顿率高",
			metrics: types.VideoQoSMetrics{
				StallRate:   5000,
				ReportCount: 10,
			},
			wantMin: 0.7,
			wantMax: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorer.qosBoost(tt.metrics)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("qosBoost() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
