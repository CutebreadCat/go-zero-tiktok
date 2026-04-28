// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/mw/token"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"go_zero-tiktok/internal/svc/xerr"
	"net/http"

	mfa_code "go_zero-tiktok/internal/mw/mfa_code"

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
		logx.Errorf("username or password is empty")
		return nil, xerr.New(http.StatusBadRequest, "用户名或密码不能为空")
	}

	user, err := l.svcCtx.Dal.User.GetUserByUsername(l.ctx, req.Username)
	if err != nil {
		logx.Errorf("get user by username failed: %v", err)
		return nil, xerr.New(http.StatusBadRequest, "用户名或密码错误")
	}

	if !myutils.CompareHashAndPassword(req.Password, user.Password) {
		logx.Errorf("invalid username or password")
		return nil, xerr.New(http.StatusBadRequest, "用户名或密码错误")
	}
	var mfa_ok bool
	mfa_ok, err = l.svcCtx.Dal.User.CheckExistsMFA(l.ctx, user.UserID)
	if err != nil {
		logx.Errorf("check mfa failed: %v", err)
		return nil, xerr.New(http.StatusInternalServerError, "检查 MFA 失败")
	}
	if mfa_ok {
		logx.Infof("user %d has mfa enabled, require mfa verification", user.UserID)
		if req.MfaCode == "" {
			return nil, xerr.New(http.StatusBadRequest, "MFA 代码不能为空")
		}
		var secret string
		if secret, err = l.svcCtx.Dal.User.FindUserMFASecret(l.ctx, user.UserID); err != nil {
			logx.Errorf("find user mfa secret failed: %v", err)
			return nil, xerr.New(http.StatusInternalServerError, "获取 MFA 密钥失败")
		}
		err := mfa_code.ValidateMfaCode(l.ctx, secret, req.MfaCode)
		if err != nil {
			logx.Errorf("validate mfa code failed: %v", err)
			return nil, xerr.New(http.StatusInternalServerError, "验证 MFA 失败")
		}

		logx.Infof("mfa verification passed for user %d", user.UserID)

	}
	accessToken, err := token.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		logx.Errorf("generate access token failed: %v", err)
		return nil, xerr.New(http.StatusInternalServerError, "生成访问令牌失败")
	}

	refreshToken, err := token.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, user.UserID)
	if err != nil {
		logx.Errorf("generate refresh token failed: %v", err)
		return nil, xerr.New(http.StatusInternalServerError, "生成刷新令牌失败")
	}

	if err := token.SaveRefreshToken(l.ctx, l.svcCtx.Rdb, refreshToken, user.UserID); err != nil {
		logx.Errorf("save refresh token failed: %v", err)
		return nil, xerr.New(http.StatusInternalServerError, "保存刷新令牌失败")
	}

	resp = &types.LoginResponse{
		Base:         types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return
}
