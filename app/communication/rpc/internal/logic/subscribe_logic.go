package logic

import (
	"context"
	"fmt"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type SubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *SubscribeLogic) Subscribe(in *communication_pb.SubscribeRequest) (*communication_pb.SubscribeResponse, error) {
	if in.FollowerId == 0 {
		return nil, xerr.NewInvalidParam("关注者ID不能为空")
	}
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("被关注用户ID不能为空")
	}

	switch in.ActionType {
	case 1:
		if err := l.svcCtx.UserFollowService.FollowUser(l.ctx, in.FollowerId, in.UserId); err != nil {
			return nil, xerr.HandleDaoError(err, "Subscribe.FollowUser")
		}
		// 创建关注消息通知（非关键路径，失败仅记日志不影响主响应）。
		eventID := fmt.Sprintf("FOLLOW:%d:%d", in.FollowerId, in.UserId)
		if _, err := l.svcCtx.MessageService.Create(l.ctx, in.UserId, "FOLLOW", "新增关注", "有人关注了你",
			eventID, eventID, in.FollowerId, "", "", in.FollowerId, "user"); err != nil {
			l.Errorf("Subscribe.CreateFollowMessage failed follower=%d user=%d: %v", in.FollowerId, in.UserId, err)
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
