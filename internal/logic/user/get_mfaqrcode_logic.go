// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"net/http"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	mfa "go_zero-tiktok/internal/middleware/mfa"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMfaqrcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMfaqrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMfaqrcodeLogic {
	return &GetMfaqrcodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMfaqrcodeLogic) GetMfaqrcode(req *types.MfaqrcodeRequest) (resp *types.MfaqrcodeResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		logx.Errorf("failed to get user id from context, error: %v", err)
		return nil, err
	}
	secret, url, err := mfa.GenerateSecret(l.ctx, userID)
	if err != nil {
		logx.Errorf("failed to generate mfa secret for user_id: %s, error: %v", userID, err)
		return nil, err
	}
	err = l.svcCtx.Dal.User.UpdateUserMFAPendingSecret(l.ctx, userID, secret)
	if err != nil {
		logx.Errorf("failed to update user mfa pending secret for user_id: %s, error: %v", userID, err)
		return nil, err
	}
	resp = &types.MfaqrcodeResponse{
		Mfa_secret: secret,
		QRCodeURL:  url,
		Base: types.BaseResponse{
			StatusCode: http.StatusOK,
			StatusMsg:  "获取secret成功",
		},
	}
	return
}
