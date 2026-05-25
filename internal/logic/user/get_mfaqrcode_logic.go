package user

import (
	"context"
	"net/http"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
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

	secret, url, err := l.svcCtx.UserMfaService.GenerateQRCode(l.ctx, userID)
	if err != nil {
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
