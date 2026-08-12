package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
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

func (l *CancelFavoriteVideoLogic) CancelFavoriteVideo(in *video_pb.CancelFavoriteVideoRequest) (*video_pb.CancelFavoriteVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	if err := l.svcCtx.VideoService.CancelFavoriteVideo(l.ctx, in.UserId, in.VideoId); err != nil {
		return nil, xerr.HandleDaoError(err, "CancelFavoriteVideo")
	}

	return &video_pb.CancelFavoriteVideoResponse{}, nil
}
