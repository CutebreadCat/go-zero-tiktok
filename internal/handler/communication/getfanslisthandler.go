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

func GetFansListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFansListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := communication.NewGetFansListLogic(r.Context(), svcCtx)
		resp, err := l.GetFansList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
