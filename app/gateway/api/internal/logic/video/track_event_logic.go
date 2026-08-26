package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	logger "go_zero-tiktok/pkg/logger"
	myutils "go_zero-tiktok/pkg/utils"
)

type TrackEventLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTrackEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrackEventLogic {
	return &TrackEventLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *TrackEventLogic) TrackEvent(req *types.TrackingEventsRequest) (resp *types.TrackingEventResponse, err error) {
	userID, _ := myutils.GetUserIDFromContext(l.ctx)

	if err := l.svcCtx.TrackingEventProducer.SendBatch(l.ctx, userID, req.Events); err != nil {
		l.Errorf("send tracking events failed: %v", err)
		// 埋点上报失败不阻断主链路，返回成功避免客户端重试风暴
	}

	return &types.TrackingEventResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
	}, nil
}
