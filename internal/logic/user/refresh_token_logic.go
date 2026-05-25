package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	if req.RefreshToken == "" {
		return nil, xerr.NewUnauthorized("刷新令牌不能为空")
	}

	accessToken, refreshToken, err := l.svcCtx.UserAuthService.RefreshToken(l.ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	resp = &types.RefreshTokenResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return
}
