package user

import (
	"context"
	"mime/multipart"

	"go_zero-tiktok/internal/infra/storage/aliyun"
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

	objectKey := "user_photos/" + userid + "/" + "profile_photo.jpg"

	userinfo, err := l.svcCtx.Dal.User.GetUserByID(l.ctx, userid)
	if err != nil {
		return nil, err
	}

	if userinfo.PhotoURL != "" && userinfo.PhotoURL != "https://example.com/default_photo.jpg" {
		if err := aliyun.DeleteFileFromOSS(objectKey); err != nil {
			return nil, xerr.HandleDaoError(err, "PostUserPhoto.DeleteOldPhoto")
		}
	}

	photoURL, err := aliyun.UploadBytesToOSS(file, objectKey)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "PostUserPhoto.UploadToOSS")
	}

	if err := l.svcCtx.Dal.User.UpdateUserPhotoByID(l.ctx, userid, photoURL); err != nil {
		return nil, xerr.HandleDaoError(err, "PostUserPhoto.UpdateUserPhoto")
	}

	return &types.UserphotoResponse{
		StatusCode: 200,
		StatusMsg:  "照片上传成功",
	}, nil
}
