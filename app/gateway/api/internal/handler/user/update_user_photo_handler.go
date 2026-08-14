// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/app/gateway/api/internal/logic/user"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/pkg/xerr"
)

func UpdateUserPhotoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("照片上传失败"))
			return
		}

		// 通道 1: multipart/form-data 真文件上传(优先)
		var photo []byte
		if file, _, err := r.FormFile("file"); err == nil {
			photo, err = io.ReadAll(file)
			file.Close()
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("读取照片文件失败"))
				return
			}
		}

		var req types.UpdateUserPhotoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 通道 2: file 字段传 base64 文本(兜底)
		if len(photo) == 0 && req.File != "" {
			b, err := base64.StdEncoding.DecodeString(req.File)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("照片内容格式错误"))
				return
			}
			photo = b
		}

		if len(photo) == 0 {
			httpx.ErrorCtx(r.Context(), w, xerr.NewInvalidParam("未找到上传文件"))
			return
		}

		l := user.NewUpdateUserPhotoLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUserPhoto(photo)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
