// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/internal/logic/user"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
)

func PostUserPhotoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserphotoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewPostUserPhotoLogic(r.Context(), svcCtx)
		resp, err := l.PostUserPhoto(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
