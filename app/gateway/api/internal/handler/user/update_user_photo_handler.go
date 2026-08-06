// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/app/gateway/api/internal/logic/user"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/pkg/xerr"
)

func UpdateUserPhotoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("照片上传失败"))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("未找到上传文件"))
			return
		}
		defer file.Close()

		l := user.NewUpdateUserPhotoLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUserPhoto(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}