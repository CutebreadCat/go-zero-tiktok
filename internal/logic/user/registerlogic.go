// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"
	"time"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

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
		return nil, errors.New("username or password is empty")
	}

	if _, err := dal.GetUserByUsername(l.ctx, req.Username); err == nil {
		logx.Errorf("user already exists: %s", req.Username)
		return nil, errors.New("user already exists")
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

	if err := dal.CreateUser(l.ctx, user); err != nil {
		logx.Errorf("failed to create user: %v", err)

		return nil, err
	}

	resp = &types.RegisterResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID: user.UserID,
	}

	return
}
