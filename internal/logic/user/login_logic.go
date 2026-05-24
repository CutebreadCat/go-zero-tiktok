package user

import (
	"context"

	"go_zero-tiktok/internal/middleware/token"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	mfa_code "go_zero-tiktok/internal/middleware/mfa"

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

	user, err := l.svcCtx.Dal.User.GetUserByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, xerr.NewInvalidParam("用户名或密码错误")
	}

	if !myutils.CompareHashAndPassword(req.Password, user.Password) {
		return nil, xerr.NewInvalidParam("用户名或密码错误")
	}

	var mfa_ok bool
	mfa_ok, err = l.svcCtx.Dal.User.CheckExistsMFA(l.ctx, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.CheckExistsMFA")
	}
	if mfa_ok {
		if req.MfaCode == "" {
			return nil, xerr.NewInvalidParam("MFA 代码不能为空")
		}
		var secret string
		if secret, err = l.svcCtx.Dal.User.FindUserPendMFASecret(l.ctx, user.UserID); err != nil {
			return nil, xerr.HandleDaoError(err, "Login.FindUserPendMFASecret")
		}
		if err := mfa_code.ValidateMfaCode(l.ctx, secret, req.MfaCode); err != nil {
			return nil, xerr.NewInvalidParam("MFA 验证失败")
		}
	}

	accessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.GenerateAccessToken")
	}

	refreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.GenerateRefreshToken")
	}

	if err := token.SaveRefreshToken(l.ctx, l.svcCtx.Rdb, refreshToken, user.UserID); err != nil {
		return nil, xerr.HandleDaoError(err, "Login.SaveRefreshToken")
	}

	resp = &types.LoginResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return
}
