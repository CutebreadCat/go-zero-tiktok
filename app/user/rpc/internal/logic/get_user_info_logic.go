package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user_pb.GetUserInfoRequest) (*user_pb.GetUserInfoResponse, error) {
	if in.UserId == 0 {
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
