package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
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
