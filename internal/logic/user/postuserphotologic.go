// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"mime/multipart"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"go_zero-tiktok/internal/mw/ali"

	"go_zero-tiktok/internal/mw/token"

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
		logx.WithContext(l.ctx).Errorf("获取用户ID失败")
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}
	logx.WithContext(l.ctx).Infof("获取到用户ID: %s", userid)

	objectKey := "user_photos/" + userid + "/" + "profile_photo.jpg"
	logx.WithContext(l.ctx).Infof("生成的对象键: %s", objectKey)

	var userinfo *types.UserBaseinfo

	if userinfo, err = l.svcCtx.Dal.User.GetUserByID(l.ctx, userid); err != nil {
		logx.WithContext(l.ctx).Errorf("用户不存在: %v", err)
		return nil, xerr.New(404, "用户不存在，无法上传头像")
	}

	if userinfo.PhotoURL != "" && userinfo.PhotoURL != "https://example.com/default_photo.jpg" {
		logx.WithContext(l.ctx).Infof("用户已有头像，开始删除旧头像")
		if err := ali.DeleteFileFromOSS(objectKey); err != nil {
			logx.WithContext(l.ctx).Errorf("删除用户旧头像失败: %v", err)
			return nil, xerr.New(1004, "头像更新失败，请稍后重试")
		}

	}

	photoURL, err := ali.UploadBytesToOSS(file, objectKey)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("上传用户头像失败: %v", err)
		return nil, xerr.New(1004, "上传头像失败，请稍后重试")
	}

	if err := l.svcCtx.Dal.User.UpdateUserPhotoByID(l.ctx, userid, photoURL); err != nil {
		logx.WithContext(l.ctx).Errorf("数据库当中更新用户头像失败: %v", err)
		return nil, xerr.New(1002, "头像保存失败，请稍后重试")
	}

	return &types.UserphotoResponse{
		StatusCode: 200,
		StatusMsg:  "照片上传成功",
	}, nil
}
