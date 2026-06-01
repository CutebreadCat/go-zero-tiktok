package user

import (
	"context"

	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	result, err := l.svcCtx.UserRpc.Login(l.ctx, &userservice.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		MfaCode:  req.MfaCode,
	})
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID:       result.UserId,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
