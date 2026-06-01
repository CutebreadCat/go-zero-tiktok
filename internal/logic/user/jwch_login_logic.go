package user

import (
	"context"

	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return &types.JwchLoginResponse{
			Base: types.BaseResponse{StatusCode: 400, StatusMsg: "获取用户 ID 失败"},
		}, nil
	}

	if _, err := l.svcCtx.UserRpc.JwchLogin(l.ctx, &userservice.JwchLoginRequest{
		UserId:   userID,
		Username: req.Username,
		Password: req.Password,
	}); err != nil {
		return &types.JwchLoginResponse{
			Base: types.BaseResponse{StatusCode: 400, StatusMsg: err.Error()},
		}, nil
	}
	return &types.JwchLoginResponse{
		Base: types.BaseResponse{StatusCode: 200, StatusMsg: "登录成功"},
	}, nil
}
