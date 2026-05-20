package token

import (
	"context"
	"net/http"
	"strings"
)

type Claims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
}

func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value(UserIDContextKey); value != nil {
		if userID, ok := value.(string); ok {
			return userID
		}
	}

	return ""
}

func extractRefreshToken(r *http.Request) string {
	if c, err := r.Cookie(refreshTokenName); err == nil && c != nil && c.Value != "" {
		return c.Value
	}

	authorization := strings.TrimSpace(r.Header.Get(authorizationHeader))
	if authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) == 2 && strings.EqualFold(parts[0], bearerTokenPrefix) && parts[1] != "" {
			return parts[1]
		}
	}

	if err := r.ParseForm(); err == nil {
		if v := strings.TrimSpace(r.FormValue(refreshTokenName)); v != "" {
			return v
		}
	}

	return ""
}
