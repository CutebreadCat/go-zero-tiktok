package token

import (
	"context"
	"net/http"
	"strings"

	"go_zero-tiktok/internal/shared/xerr"

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
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, xerr.NewUnauthorized(tokenInvalidMessage).(*xerr.CodeError).HandleResponse())
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
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized(tokenInvalidMessage))
				return
			}

			claims, err := ParseToken(secret, accessToken)
			if err != nil || claims.Claims.TokenType != AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized(tokenInvalidMessage))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.Claims.UserID)
			next(w, r.WithContext(ctx))
		}
	}
}

func isPublicPath(path string) bool {
	_, ok := publicPaths[path]
	return ok
}

func extractAccessToken(r *http.Request) string {
	if authorization := r.Header.Get(authorizationHeader); authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) == 2 && strings.EqualFold(parts[0], bearerTokenPrefix) {
			return parts[1]
		}
	}

	return ""
}
