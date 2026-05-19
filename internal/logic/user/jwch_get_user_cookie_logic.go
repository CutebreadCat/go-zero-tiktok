// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/west2-online/jwch"
	"github.com/zeromicro/go-zero/core/logx"
)

type JwchGetUserCookieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJwchGetUserCookieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwchGetUserCookieLogic {
	return &JwchGetUserCookieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JwchGetUserCookieLogic) JwchGetUserCookie(req *types.JwchGetUserCookieRequest) (resp *types.JwchGetUserCookieResponse, err error) {
	// 获取当前登录用户ID
	userid, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "获取用户ID失败",
			},
		}, nil
	}

	// 从数据库获取用户的教务处信息
	jwchID, jwchPassword, err := l.svcCtx.Dal.User.GetUserJwchInfo(l.ctx, userid)
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "获取教务处信息失败，请先绑定教务处账号",
			},
		}, nil
	}

	// 教务处登录
	jwchClient := jwch.NewStudent()
	jwchClient.ID = jwchID
	jwchClient.Password = jwchPassword
	err = jwchClient.Login()
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "教务处登录失败，请检查账号密码是否正确",
			},
		}, nil
	}

	// 获取用户标识和cookie
	user, cookie, err := jwchClient.GetIdentifierAndCookies()
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "获取cookie失败",
			},
		}, nil
	}

	jwchcookie := myutils.ParseCookieTostring(cookie)

	return &types.JwchGetUserCookieResponse{
		Base: types.BaseResponse{
			StatusCode: 200,
			StatusMsg:  "获取成功",
		},
		JwchID: user,
		Cookie: jwchcookie,
	}, nil
}
