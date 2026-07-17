package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFansListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFansListLogic {
	return &GetFansListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFansListLogic) GetFansList(in *communication_pb.GetFansListRequest) (*communication_pb.GetFansListResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	users, total, err := l.svcCtx.UserFollowService.GetFansList(l.ctx, in.UserId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFansList")
	}

	userInfos := make([]*communication_pb.UserInfo, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, &communication_pb.UserInfo{
			UserId:   u.UserID,
			Username: u.Username,
			PhotoUrl: u.PhotoURL,
		})
	}

	return &communication_pb.GetFansListResponse{
		Users: userInfos,
		Total: total,
	}, nil
}
