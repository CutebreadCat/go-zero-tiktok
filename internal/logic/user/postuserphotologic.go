// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostUserPhotoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostUserPhotoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostUserPhotoLogic {
	return &PostUserPhotoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PostUserPhotoLogic) PostUserPhoto(req *types.UserphotoRequest) (resp *types.UserphotoResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
