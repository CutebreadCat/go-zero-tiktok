package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
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

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	result, err := l.svcCtx.UserAuthService.Login(l.ctx, req.Username, req.Password, req.MfaCode)
	if err != nil {
		return nil, err
	}

	resp = &types.LoginResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	return
}
