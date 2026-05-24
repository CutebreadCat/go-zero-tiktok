package user

import (
	"context"
	"net/http"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
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
		return nil, xerr.NewUnauthorized("获取用户ID失败")
	}
	secret, url, err := mfa.GenerateSecret(l.ctx, userID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetMfaqrcode.GenerateSecret")
	}
	err = l.svcCtx.Dal.User.UpdateUserMFAPendingSecret(l.ctx, userID, secret)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetMfaqrcode.UpdateUserMFAPendingSecret")
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
