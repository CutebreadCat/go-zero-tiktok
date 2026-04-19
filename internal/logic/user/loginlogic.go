// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/mw/token"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

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
		return nil, errors.New("username or password is empty")
	}

	user, err := dal.GetUserByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if !myutils.CompareHashAndPassword(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	accessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		return nil, err
	}

	if err := token.SaveRefreshToken(l.ctx, l.svcCtx.Rdb, refreshToken, user.UserID); err != nil {
		return nil, err
	}

	resp = &types.LoginResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return
}
