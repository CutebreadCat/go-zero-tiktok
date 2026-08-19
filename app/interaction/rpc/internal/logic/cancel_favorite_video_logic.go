package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CancelFavoriteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewCancelFavoriteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelFavoriteVideoLogic {
	return &CancelFavoriteVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *CancelFavoriteVideoLogic) CancelFavoriteVideo(in *interaction_pb.CancelFavoriteVideoRequest) (*interaction_pb.CancelFavoriteVideoResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	if err := l.svcCtx.InteractionService.CancelFavoriteVideo(l.ctx, in.UserId, in.VideoId); err != nil {
		return nil, xerr.HandleDaoError(err, "CancelFavoriteVideo")
	}

	return &interaction_pb.CancelFavoriteVideoResponse{}, nil
}
