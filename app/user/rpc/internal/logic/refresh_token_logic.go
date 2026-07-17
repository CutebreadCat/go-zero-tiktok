package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *user_pb.RefreshTokenRequest) (*user_pb.RefreshTokenResponse, error) {
	if in.RefreshToken == "" {
		return nil, xerr.NewInvalidParam("刷新令牌不能为空")
	}

	accessToken, refreshToken, err := l.svcCtx.UserAuthService.RefreshToken(l.ctx, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &user_pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
