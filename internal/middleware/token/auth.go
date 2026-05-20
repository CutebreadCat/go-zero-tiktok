package token

import (
	"context"
	"net/http"
	"strings"

	"go_zero-tiktok/internal/svc/xerr"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func WithAuth(secret string) rest.RunOption {
	return func(server *rest.Server) {
		rest.WithUnauthorizedCallback(UnauthorizedCallback)(server)
		server.Use(AuthMiddleware(secret))
	}
}

func UnauthorizedCallback(w http.ResponseWriter, r *http.Request, err error) {
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, xerr.NewUnauthorized("token无效").(*xerr.CodeError).HandleResponse())
}

func AuthMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next(w, r)
				return
			}

			accessToken := extractAccessToken(r)
			if accessToken == "" {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token无效"))
				return
			}

			claims, err := ParseToken(secret, accessToken)
			if err != nil || claims.Claims.TokenType != AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token无效"))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.Claims.UserID)
			next(w, r.WithContext(ctx))
		}
	}
}

func isPublicPath(path string) bool {
	switch path {
	case "/user/login",
		"/user/register",
		"/user/token/refresh",
		"/video/list",
		"/video/popular",
		"/video/search":
		return true
	default:
		return false
	}
}

func extractAccessToken(r *http.Request) string {
	if authorization := r.Header.Get("Authorization"); authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
}
