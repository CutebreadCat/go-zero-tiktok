// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/mw/token"
	"go_zero-tiktok/internal/svc"
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
		return nil, errors.New("refresh token is empty")
	}

	claims, err := token.ParseToken(l.svcCtx.Config.Auth.AccessSecret, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if claims.Claims.TokenType != token.RefreshTokenType {
		return nil, errors.New("invalid refresh token")
	}

	userID, err := token.GetRefreshTokenUserID(l.ctx, l.svcCtx.Rdb, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if userID != claims.Claims.UserID {
		return nil, errors.New("refresh token mismatch")
	}

	newAccessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, userID)
	if err != nil {
		return nil, err
	}

	if err := token.RotateRefreshToken(l.ctx, l.svcCtx.Rdb, req.RefreshToken, newRefreshToken, userID); err != nil {
		return nil, err
	}

	resp = &types.RefreshTokenResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return
}
