// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/app/gateway/api/internal/logic/communication"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
)

func MarkReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MarkReadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := communication.NewMarkReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkRead(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
