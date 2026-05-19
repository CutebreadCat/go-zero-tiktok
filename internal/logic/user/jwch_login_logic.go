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

type JwchLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJwchLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwchLoginLogic {
	return &JwchLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JwchLoginLogic) JwchLogin(req *types.JwchLoginRequest) (resp *types.JwchLoginResponse, err error) {
	// todo: add your logic here and delete this line
	jwchClient := jwch.NewStudent()
	jwchClient.ID = req.Username
	jwchClient.Password = req.Password

	err = jwchClient.Login()
	if err != nil {
		return &types.JwchLoginResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "你的账号密码无法通过教务处认证，请检查后重试",
			},
		}, nil
	}
	// 登录成功，更新数据库中的用户信息
	var userid string
	userid, err = myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return &types.JwchLoginResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "获取用户ID失败",
			},
		}, nil
	}
	err = l.svcCtx.Dal.User.UpdateUserJwchInfo(l.ctx, userid, req.Username, req.Password)
	if err != nil {
		return &types.JwchLoginResponse{
			Base: types.BaseResponse{
				StatusCode: 400,
				StatusMsg:  "更新用户信息失败",
			},
		}, nil
	}
	return &types.JwchLoginResponse{
		Base: types.BaseResponse{
			StatusCode: 200,
			StatusMsg:  "登录成功",
		},
	}, nil
}
