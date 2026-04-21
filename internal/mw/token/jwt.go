package token

import (
	"time"

	"go_zero-tiktok/internal/svc/xerr"

	"github.com/golang-jwt/jwt/v4"
)

type JwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

func GenerateToken(secret, userID, tokenType string, expire time.Duration) (string, error) {
	claims := JwtClaims{
		Claims: Claims{
			UserID:    userID,
			TokenType: tokenType,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret, tokenString string) (*JwtClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, xerr.New(400, "unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*JwtClaims)
	if !ok || !parsedToken.Valid {
		return nil, xerr.New(400, "invalid token")
	}

	return claims, nil
}

func GenerateAccessToken(secret, userID string) (string, error) {
	return GenerateToken(secret, userID, AccessTokenType, time.Hour)
}

func GenerateRefreshToken(secret, userID string) (string, error) {
	return GenerateToken(secret, userID, RefreshTokenType, 24*time.Hour)
}
