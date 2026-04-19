package token

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			accessToken := extractAccessToken(r)
			if accessToken == "" {
				httpx.ErrorCtx(r.Context(), w, http.ErrNoCookie)
				return
			}

			claims, err := ParseToken(secret, accessToken)
			if err != nil || claims.Claims.TokenType != AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.Claims.UserID)
			next(w, r.WithContext(ctx))
		}
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
