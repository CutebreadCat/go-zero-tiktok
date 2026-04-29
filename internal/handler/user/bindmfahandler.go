// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"

	"go_zero-tiktok/internal/logic/user"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func BindMfaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindMfaqrcodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			logx.Errorf("parse bind mfa request failed: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewBindMfaLogic(r.Context(), svcCtx)
		resp, err := l.BindMfa(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
