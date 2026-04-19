// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/internal/logic/video"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
)

func VideoPopularHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VideoPopularRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := video.NewVideoPopularLogic(r.Context(), svcCtx)
		resp, err := l.VideoPopular(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
