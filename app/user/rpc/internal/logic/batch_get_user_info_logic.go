package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type BatchGetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewBatchGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserInfoLogic {
	return &BatchGetUserInfoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *BatchGetUserInfoLogic) BatchGetUserInfo(in *user_pb.BatchGetUserInfoRequest) (*user_pb.BatchGetUserInfoResponse, error) {
	if len(in.UserIds) == 0 {
		return nil, xerr.NewInvalidParam("用户 ID 列表不能为空")
	}

	users, err := l.svcCtx.UserProfileService.GetUsersByIDs(l.ctx, in.UserIds)
	if err != nil {
		return nil, err
	}

	resp := &user_pb.BatchGetUserInfoResponse{
		Users: make([]*user_pb.UserInfo, 0, len(users)),
	}
	for i := range users {
		u := &users[i]
		resp.Users = append(resp.Users, &user_pb.UserInfo{
			UserId:    u.UserID,
			Username:  u.Username,
			PhotoUrl:  u.PhotoURL,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	return resp, nil
}
