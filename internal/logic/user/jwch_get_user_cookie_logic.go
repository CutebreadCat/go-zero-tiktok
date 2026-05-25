// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
	userid, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "获取用户ID失败",
			},
		}, nil
	}

	identifier, cookie, err := l.svcCtx.UserJwchService.GetCookie(l.ctx, userid)
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  err.Error(),
			},
		}, nil
	}

	return &types.JwchGetUserCookieResponse{
		Base: types.BaseResponse{
			StatusCode: 200,
			StatusMsg:  "获取成功",
		},
		JwchID: identifier,
		Cookie: cookie,
	}, nil
}
