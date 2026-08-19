package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type LikeVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *LikeVideoLogic) LikeVideo(in *interaction_pb.LikeVideoRequest) (*interaction_pb.LikeVideoResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	switch in.ActionType {
	case 1:
		if err := l.svcCtx.InteractionService.LikeVideo(l.ctx, in.UserId, in.VideoId); err != nil {
			return nil, xerr.HandleDaoError(err, "LikeVideo")
		}
	case 0:
		if err := l.svcCtx.InteractionService.CancelLikeVideo(l.ctx, in.UserId, in.VideoId); err != nil {
			return nil, xerr.HandleDaoError(err, "LikeVideo")
		}
	default:
		return nil, xerr.NewInvalidParam("无效的点赞动作类型")
	}

	return &interaction_pb.LikeVideoResponse{}, nil
}
