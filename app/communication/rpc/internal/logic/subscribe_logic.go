package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubscribeLogic) Subscribe(in *communication_pb.SubscribeRequest) (*communication_pb.SubscribeResponse, error) {
	if in.FollowerId == "" {
		return nil, xerr.NewInvalidParam("关注者ID不能为空")
	}
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("被关注用户ID不能为空")
	}

	switch in.ActionType {
	case 1:
		if err := l.svcCtx.UserFollowService.FollowUser(l.ctx, in.FollowerId, in.UserId); err != nil {
			return nil, xerr.HandleDaoError(err, "Subscribe.FollowUser")
		}
	case 0:
		if err := l.svcCtx.UserFollowService.UnfollowUser(l.ctx, in.FollowerId, in.UserId); err != nil {
			return nil, xerr.HandleDaoError(err, "Subscribe.UnfollowUser")
		}
	default:
		return nil, xerr.NewInvalidParam("操作类型无效，仅支持1(关注)或0(取消关注)")
	}

	return &communication_pb.SubscribeResponse{}, nil
}
