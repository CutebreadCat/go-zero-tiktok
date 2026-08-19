package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type RecalculateHotScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewRecalculateHotScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecalculateHotScoreLogic {
	return &RecalculateHotScoreLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *RecalculateHotScoreLogic) RecalculateHotScore(in *video_pb.RecalculateHotScoreRequest) (*video_pb.RecalculateHotScoreResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频 ID 不能为空")
	}

	if err := l.svcCtx.VideoService.RecalculateHotScore(l.ctx, in.VideoId); err != nil {
		return nil, err
	}

	return &video_pb.RecalculateHotScoreResponse{}, nil
}
