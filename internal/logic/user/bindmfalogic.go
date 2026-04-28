// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	mfa "go_zero-tiktok/internal/mw/mfa_code"

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
	var userId string
	if userId, err = myutils.GetUserIDFromContext(l.ctx); err != nil {
		logx.Errorf("failed to get user id from context, error: %v", err)
		return nil, err
	}
	if err = mfa.ValidateMfaCode(l.ctx, req.Mfa_secret, req.Mfa_code); err != nil {
		logx.Errorf("failed to verify mfa for user_id: %s, error: %v", userId, err)
		return nil, err
	}
	if err = l.svcCtx.Dal.User.EnableUserMFA(l.ctx, userId); err != nil {
		logx.Errorf("failed to update mfa secret for user_id: %s, error: %v", userId, err)
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
