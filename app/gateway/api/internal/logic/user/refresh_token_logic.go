package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type RefreshTokenLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, xerr.NewUnauthorized("刷新令牌不能为空")
	}

	result, err := l.svcCtx.UserRpc.RefreshToken(l.ctx, &userservice.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	return &types.RefreshTokenResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
