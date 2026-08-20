package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type ReportPlaybackQoSLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportPlaybackQoSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportPlaybackQoSLogic {
	return &ReportPlaybackQoSLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *ReportPlaybackQoSLogic) ReportPlaybackQoS(req *types.PlaybackQoSReportRequest) (resp *types.PlaybackQoSReportResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	_, err = l.svcCtx.VideoRpc.ReportPlaybackQoS(l.ctx, &videopb.PlaybackQoSReportRequest{
		UserId:         userID,
		VideoId:        req.VideoID,
		IdempotencyKey: req.IdempotencyKey,
		EventType:      req.EventType,
		DurationMs:     req.DurationMs,
		PlayedMs:       req.PlayedMs,
		BufferedMs:     req.BufferedMs,
		StallCount:     req.StallCount,
		StallTotalMs:   req.StallTotalMs,
		Resolution:     req.Resolution,
		BitrateKbps:    req.BitrateKbps,
		Fps:            req.Fps,
		ErrorCode:      req.ErrorCode,
		ErrorMsg:       req.ErrorMsg,
		NetworkType:    req.NetworkType,
		DeviceInfo:     req.DeviceInfo,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "ReportPlaybackQoS")
	}

	return &types.PlaybackQoSReportResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
	}, nil
}
