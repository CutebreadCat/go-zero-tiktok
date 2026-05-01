package token

import (
	"context"
	"net/http"
	"strings"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
	UserIDContextKey = "user_id"
	RefreshPrefix    = "refresh_token:"
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
	if c, err := r.Cookie("refresh_token"); err == nil && c != nil && c.Value != "" {
		return c.Value
	}

	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
			return parts[1]
		}
	}

	if err := r.ParseForm(); err == nil {
		if v := strings.TrimSpace(r.FormValue("refresh_token")); v != "" {
			return v
		}
	}

	return ""
}
