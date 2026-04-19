// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"
	"time"

	"go_zero-tiktok/internal/logic/user"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewRefreshTokenLogic(r.Context(), svcCtx)
		resp, err := l.RefreshToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if resp.RefreshToken != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "refresh_token",
					Value:    resp.RefreshToken,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   int((24 * time.Hour).Seconds()),
				})
				resp.RefreshToken = ""
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
