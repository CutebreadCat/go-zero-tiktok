// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"time"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
		logx.Error("username or password is empty")
		return nil, xerr.New(400, "用户名或密码不能为空")
	}

	if _, err := l.svcCtx.Dal.User.GetUserByUsername(l.ctx, req.Username); err == nil {
		logx.Errorf("user already exists: %s", req.Username)
		return nil, xerr.New(409, "用户名已存在，请更换后重试")
	}

	user := &types.UserBaseinfo{
		UserID:    myutils.GenerateUserID(),
		Username:  req.Username,
		Password:  myutils.HashPassword(req.Password),
		PhotoURL:  "",
		CreatedAt: myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		UpdatedAt: myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		DeletedAt: "",
	}

	if err := l.svcCtx.Dal.User.CreateUser(l.ctx, user); err != nil {
		logx.Errorf("failed to create user: %v", err)

		return nil, xerr.New(1002, "注册失败，请稍后重试")
	}

	resp = &types.RegisterResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID: user.UserID,
	}

	return
}
