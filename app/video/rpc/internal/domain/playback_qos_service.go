package domain

import (
	"context"
	"strings"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"
)

// PlaybackQoSService 播放质量上报领域服务。
// 负责参数校验、幂等去重与落库。
type PlaybackQoSService struct {
	playbackQoSRepo IPlaybackQoSRepo
}

// NewPlaybackQoSService 创建播放质量上报领域服务。
func NewPlaybackQoSService(playbackQoSRepo IPlaybackQoSRepo) *PlaybackQoSService {
	return &PlaybackQoSService{
		playbackQoSRepo: playbackQoSRepo,
	}
}

// validEventTypes 合法的事件类型集合。
var validEventTypes = map[string]struct{}{
	"start":    {},
	"playing":  {},
	"pause":    {},
	"complete": {},
	"error":    {},
	"stall":    {},
}

// ReportPlaybackQoS 处理播放质量上报请求。
// 对重复上报（同一 user_id + idempotency_key）返回成功，不重复写库。
func (s *PlaybackQoSService) ReportPlaybackQoS(ctx context.Context, report *types.PlaybackQoSReport) error {
	if report == nil {
		return xerr.NewInvalidParam("上报数据为空")
	}
	if report.UserID <= 0 {
		return xerr.NewInvalidParam("用户 ID 无效")
	}
	if report.VideoID <= 0 {
		return xerr.NewInvalidParam("视频 ID 无效")
	}
	if strings.TrimSpace(report.IdempotencyKey) == "" {
		return xerr.NewInvalidParam("幂等键不能为空")
	}
	if _, ok := validEventTypes[report.EventType]; !ok {
		return xerr.NewInvalidParam("事件类型无效")
	}

	if err := s.playbackQoSRepo.CreateReport(ctx, report); err != nil {
		return xerr.Wrap(err, "ReportPlaybackQoS")
	}
	return nil
}
