package token

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"go_zero-tiktok/pkg/ctxkey"
	jwtpkg "go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/xerr"

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
	var codeErr *xerr.CodeError
	if !errors.As(xerr.NewUnauthorized("token invalid"), &codeErr) {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, codeErr.HandleResponse())
}

func AuthMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, ok := publicPaths[r.URL.Path]; ok {
				next(w, r)
				return
			}

			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			claims, err := jwtpkg.ParseToken(secret, parts[1])
			if err != nil || claims.TokenType != jwtpkg.AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.UserID, claims.UserID)
			next(w, r.WithContext(ctx))
		}
	}
}

var publicPaths = map[string]struct{}{
	"/user/login":         {},
	"/user/register":      {},
	"/user/token/refresh": {},
	"/video/list":         {},
	"/video/popular":      {},
	"/video/search":       {},
}

func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value(ctxkey.UserID).(string)
	return userID
}
