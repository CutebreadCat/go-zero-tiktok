// Code scaffolded by goctl. Safe to edit.

package user

import (
	"net/http"

	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/app/gateway/api/internal/logic/user"
	"go_zero-tiktok/app/gateway/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PostUserPhotoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("photo_url")
		if err != nil {
			appLogger.Errorf("获取文件失败: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := user.NewPostUserPhotoLogic(r.Context(), svcCtx)
		resp, err := l.PostUserPhoto(nil, file)
		if err != nil {
			appLogger.Errorf("上传头像失败: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		appLogger.Info("上传头像成功")
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
