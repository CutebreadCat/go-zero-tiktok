package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type IncreaseVideoVisitCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewIncreaseVideoVisitCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncreaseVideoVisitCountLogic {
	return &IncreaseVideoVisitCountLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *IncreaseVideoVisitCountLogic) IncreaseVideoVisitCount(in *video_pb.IncreaseVideoVisitCountRequest) (*video_pb.IncreaseVideoVisitCountResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频 ID 不能为空")
	}

	delta := in.Delta
	if delta == 0 {
		delta = 1
	}

	if err := l.svcCtx.VideoService.IncreaseVideoVisitCount(l.ctx, in.VideoId, delta); err != nil {
		return nil, err
	}

	return &video_pb.IncreaseVideoVisitCountResponse{}, nil
}
