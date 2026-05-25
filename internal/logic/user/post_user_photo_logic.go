package user

import (
	"context"
	"mime/multipart"

	"go_zero-tiktok/internal/middleware/token"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
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

func (l *PostUserPhotoLogic) PostUserPhoto(req *types.UserphotoRequest, file multipart.File) (resp *types.UserphotoResponse, err error) {
	userid := token.UserIDFromContext(l.ctx)
	if userid == "" {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	if err := l.svcCtx.UserProfileService.UpdatePhoto(l.ctx, userid, file); err != nil {
		return nil, err
	}

	return &types.UserphotoResponse{
		StatusCode: 200,
		StatusMsg:  "照片上传成功",
	}, nil
}
