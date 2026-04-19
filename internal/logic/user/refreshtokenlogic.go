// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/mw/token"
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
		logx.Error("refresh token is empty")
		return nil, xerr.New(1005, "刷新令牌不能为空")
	}

	claims, err := token.ParseToken(l.svcCtx.Config.Auth.AccessSecret, req.RefreshToken)
	if err != nil {
		logx.Errorf("failed to parse refresh token: %v", err)
		return nil, xerr.New(1005, "解析刷新令牌失败")
	}
	if claims.Claims.TokenType != token.RefreshTokenType {
		logx.Error("invalid token type")
		return nil, xerr.New(1005, "无效的刷新令牌")
	}

	userID, err := token.GetRefreshTokenUserID(l.ctx, l.svcCtx.Rdb, req.RefreshToken)
	if err != nil {
		logx.Errorf("failed to get user ID from refresh token: %v", err)
		return nil, xerr.New(1005, "获取用户ID失败")
	}
	if userID != claims.Claims.UserID {
		logx.Error("refresh token user ID mismatch")
		return nil, xerr.New(1005, "刷新令牌不匹配")
	}

	newAccessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		logx.Errorf("failed to generate new access token: %v", err)
		return nil, xerr.New(1005, "生成访问令牌失败")
	}

	newRefreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		logx.Errorf("failed to generate new refresh token: %v", err)

		return nil, xerr.New(1005, "生成刷新令牌失败")
	}

	if err := token.RotateRefreshToken(l.ctx, l.svcCtx.Rdb, req.RefreshToken, newRefreshToken, userID); err != nil {
		logx.Errorf("failed to rotate refresh token: %v", err)
		return nil, xerr.New(1005, "旋转刷新令牌失败")
	}

	resp = &types.RefreshTokenResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return
}
