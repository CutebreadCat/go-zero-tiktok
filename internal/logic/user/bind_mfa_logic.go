package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMfaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindMfaLogic) BindMfa(req *types.BindMfaqrcodeRequest) (resp *types.BindMfaqrcodeResponse, err error) {
	userId, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户ID失败")
	}

	if err := l.svcCtx.UserMfaService.BindMFA(l.ctx, userId, req.Mfa_secret, req.Mfa_code); err != nil {
		return nil, err
	}

	resp = &types.BindMfaqrcodeResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "绑定MFA成功",
		},
	}

	return resp, nil
}
