package token

import (
	"go_zero-tiktok/pkg/xerr"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
}

type JwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

// ParseToken 校验并解析 access token，供网关内部认证使用。
func ParseToken(secret, tokenString string) (*JwtClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, xerr.New(400, unexpectedSigningMethodMessage)
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*JwtClaims)
	if !ok || !parsedToken.Valid {
		return nil, xerr.New(400, invalidTokenMessage)
	}

	return claims, nil
}
