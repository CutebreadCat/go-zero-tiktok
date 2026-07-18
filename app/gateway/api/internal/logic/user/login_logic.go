package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type LoginLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
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
