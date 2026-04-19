package token

import "context"

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
