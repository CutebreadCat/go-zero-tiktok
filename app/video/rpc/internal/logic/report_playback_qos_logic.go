package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type ReportPlaybackQoSLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewReportPlaybackQoSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportPlaybackQoSLogic {
	return &ReportPlaybackQoSLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *ReportPlaybackQoSLogic) ReportPlaybackQoS(in *video_pb.PlaybackQoSReportRequest) (*video_pb.PlaybackQoSReportResponse, error) {
	if in == nil {
		return nil, xerr.NewInvalidParam("请求不能为空")
	}

	report := &types.PlaybackQoSReport{
		UserID:         in.UserId,
		VideoID:        in.VideoId,
		IdempotencyKey: in.IdempotencyKey,
		EventType:      in.EventType,
		DurationMs:     in.DurationMs,
		PlayedMs:       in.PlayedMs,
		BufferedMs:     in.BufferedMs,
		StallCount:     in.StallCount,
		StallTotalMs:   in.StallTotalMs,
		Resolution:     in.Resolution,
		BitrateKbps:    in.BitrateKbps,
		Fps:            in.Fps,
		ErrorCode:      in.ErrorCode,
		ErrorMsg:       in.ErrorMsg,
		NetworkType:    in.NetworkType,
		DeviceInfo:     in.DeviceInfo,
	}

	if err := l.svcCtx.PlaybackQoSService.ReportPlaybackQoS(l.ctx, report); err != nil {
		return nil, err
	}

	return &video_pb.PlaybackQoSReportResponse{}, nil
}
