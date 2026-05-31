package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user_pb.LoginRequest) (*user_pb.LoginResponse, error) {
	if in.Username == "" || in.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	result, err := l.svcCtx.UserAuthService.Login(l.ctx, in.Username, in.Password, in.MfaCode)
	if err != nil {
		return nil, err
	}

	return &user_pb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		UserId:       result.UserID,
	}, nil
}
