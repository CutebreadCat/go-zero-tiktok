package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type FavoriteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewFavoriteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteVideoLogic {
	return &FavoriteVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *FavoriteVideoLogic) FavoriteVideo(in *interaction_pb.FavoriteVideoRequest) (*interaction_pb.FavoriteVideoResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	if err := l.svcCtx.InteractionService.FavoriteVideo(l.ctx, in.UserId, in.VideoId); err != nil {
		return nil, xerr.HandleDaoError(err, "FavoriteVideo")
	}

	return &interaction_pb.FavoriteVideoResponse{}, nil
}
