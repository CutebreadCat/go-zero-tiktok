package logic

import (
	"context"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedFanoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFeedFanoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedFanoutLogic {
	return &FeedFanoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FeedFanoutLogic) FeedFanout(in *video_pb.FeedFanoutRequest) (*video_pb.FeedFanoutResponse, error) {
	if in.VideoId <= 0 {
		return nil, xerr.NewInvalidParam("video_id 不能为空")
	}
	if len(in.UserIds) == 0 {
		// 无粉丝或粉丝列表为空，无需扇出
		return &video_pb.FeedFanoutResponse{}, nil
	}

	// 扇出失败不阻断发布主流程，但向上层返回错误，便于 gateway 记录日志与监控
	if err := l.svcCtx.VideoService.FanoutToUsers(l.ctx, in.VideoId, in.UserIds, time.UnixMilli(in.PublishAt)); err != nil {
		return nil, xerr.Wrap(err, "FeedFanout")
	}
	return &video_pb.FeedFanoutResponse{}, nil
}
