package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
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

func (l *LikeVideoLogic) LikeVideo(in *video_pb.LikeVideoRequest) (*video_pb.LikeVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	switch in.ActionType {
	case 1:
		if err := l.svcCtx.VideoService.LikeVideo(l.ctx, in.UserId, in.VideoId); err != nil {
			return nil, xerr.HandleDaoError(err, "LikeVideo")
		}
	case 0:
		if err := l.svcCtx.VideoService.CancelLikeVideo(l.ctx, in.UserId, in.VideoId); err != nil {
			return nil, xerr.HandleDaoError(err, "CancelLikeVideo")
		}
	default:
		return nil, xerr.NewInvalidParam("操作类型无效，仅支持1(点赞)或0(取消点赞)")
	}

	// 记录访问量
	l.svcCtx.VideoService.RecordVisit(in.VideoId)

	return &video_pb.LikeVideoResponse{}, nil
}
