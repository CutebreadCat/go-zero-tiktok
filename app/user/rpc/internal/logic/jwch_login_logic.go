package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type JwchLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJwchLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwchLoginLogic {
	return &JwchLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *JwchLoginLogic) JwchLogin(in *user_pb.JwchLoginRequest) (*user_pb.JwchLoginResponse, error) {
	if in.UserId == "" || in.Username == "" || in.Password == "" {
		return nil, xerr.NewInvalidParam("教务处登录参数不能为空")
	}

	if err := l.svcCtx.UserJwchService.Login(l.ctx, in.UserId, in.Username, in.Password); err != nil {
		return nil, err
	}

	return &user_pb.JwchLoginResponse{}, nil
}
