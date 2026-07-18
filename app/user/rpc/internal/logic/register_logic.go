package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user_pb.RegisterRequest) (*user_pb.RegisterResponse, error) {
	if in.Username == "" || in.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	userID, err := l.svcCtx.UserAuthService.Register(l.ctx, in.Username, in.Password)
	if err != nil {
		return nil, err
	}

	return &user_pb.RegisterResponse{
		UserId: userID,
	}, nil
}
