// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"log"
	"net/http"

	"go_zero-tiktok/internal/logic/user"
	"go_zero-tiktok/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PostUserPhotoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		File, _, err := r.FormFile("photo_url")
		if err != nil {
			log.Printf("获取文件失败：%v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer File.Close()

		l := user.NewPostUserPhotoLogic(r.Context(), svcCtx)
		resp, err := l.PostUserPhoto(nil, File)
		if err != nil {
			log.Printf("上传头像失败：%v", err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			log.Printf("上传头像成功")
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
