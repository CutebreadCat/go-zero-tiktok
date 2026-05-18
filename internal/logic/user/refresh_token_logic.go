package user

import (
	"context"

	"go_zero-tiktok/internal/middleware/token"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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

	claims, err := token.ParseToken(l.svcCtx.Config.Auth.AccessSecret, req.RefreshToken)
	if err != nil {
		return nil, xerr.NewUnauthorized("解析刷新令牌失败")
	}
	if claims.Claims.TokenType != token.RefreshTokenType {
		return nil, xerr.NewUnauthorized("无效的刷新令牌")
	}

	userID, err := token.GetRefreshTokenUserID(l.ctx, l.svcCtx.Rdb, req.RefreshToken)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户ID失败")
	}
	if userID != claims.Claims.UserID {
		return nil, xerr.NewUnauthorized("刷新令牌不匹配")
	}

	newAccessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "RefreshToken.GenerateAccessToken")
	}

	newRefreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "RefreshToken.GenerateRefreshToken")
	}

	if err := token.RotateRefreshToken(l.ctx, l.svcCtx.Rdb, req.RefreshToken, newRefreshToken, userID); err != nil {
		return nil, xerr.HandleDaoError(err, "RefreshToken.RotateRefreshToken")
	}

	resp = &types.RefreshTokenResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return
}
