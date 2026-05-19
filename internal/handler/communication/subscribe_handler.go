// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/internal/logic/communication"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
)

func SubscribeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubscribeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := communication.NewSubscribeLogic(r.Context(), svcCtx)
		resp, err := l.Subscribe(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
