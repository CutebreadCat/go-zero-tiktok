package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetFriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetFriendListLogic) GetFriendList(in *communication_pb.GetFriendListRequest) (*communication_pb.GetFriendListResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	users, total, err := l.svcCtx.UserFollowService.GetFriendList(l.ctx, in.UserId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFriendList")
	}

	userInfos := make([]*communication_pb.UserInfo, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, &communication_pb.UserInfo{
			UserId:   u.UserID,
			Username: u.Username,
			PhotoUrl: u.PhotoURL,
		})
	}

	return &communication_pb.GetFriendListResponse{
		Users: userInfos,
		Total: total,
	}, nil
}
