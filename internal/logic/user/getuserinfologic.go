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

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	userID := token.UserIDFromContext(l.ctx)
	if userID == "" {
		userID = req.UserID
	}
	if userID == "" {
		return nil, errors.New("user id is empty")
	}

	user, err := dal.GetUserByID(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	resp = &types.UserInfoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		User: *user,
	}

	return
}
