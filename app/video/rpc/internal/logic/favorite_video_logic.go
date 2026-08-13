package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
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

func (l *FavoriteVideoLogic) FavoriteVideo(in *video_pb.FavoriteVideoRequest) (*video_pb.FavoriteVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	if err := l.svcCtx.InteractionService.FavoriteVideo(l.ctx, in.UserId, in.VideoId); err != nil {
		return nil, xerr.HandleDaoError(err, "FavoriteVideo")
	}

	// 记录访问量
	l.svcCtx.VideoService.RecordVisit(in.VideoId)

	return &video_pb.FavoriteVideoResponse{}, nil
}
