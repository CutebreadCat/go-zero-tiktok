package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user_pb.GetUserInfoRequest) (*user_pb.GetUserInfoResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户 ID 不能为空")
	}

	user, err := l.svcCtx.UserProfileService.GetUserByID(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &user_pb.GetUserInfoResponse{
		User: &user_pb.UserInfo{
			UserId:    user.UserID,
			Username:  user.Username,
			PhotoUrl:  user.PhotoURL,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}
