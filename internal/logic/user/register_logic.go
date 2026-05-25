package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	userID, err := l.svcCtx.UserAuthService.Register(l.ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	resp = &types.RegisterResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID: userID,
	}

	return
}
