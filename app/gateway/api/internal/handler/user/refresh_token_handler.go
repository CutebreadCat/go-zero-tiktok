// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go_zero-tiktok/app/gateway/api/internal/logic/user"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	jwtpkg "go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/xerr"
)

func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		refreshToken, err := jwtpkg.GetRefreshTokenFromCookie(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("刷新令牌不能为空"))
			return
		}
		req.RefreshToken = refreshToken

		l := user.NewRefreshTokenLogic(r.Context(), svcCtx)
		resp, err := l.RefreshToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			jwtpkg.SetAccessTokenCookie(w, resp.AccessToken)
			jwtpkg.SetRefreshTokenCookie(w, resp.RefreshToken)
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
