package token

import "time"

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
	UserIDContextKey = "user_id"
	RefreshPrefix    = "refresh_token:"

	refreshTokenName    = "refresh_token"
	authorizationHeader = "Authorization"
	bearerTokenPrefix   = "Bearer"

	tokenInvalidMessage            = "token无效"
	unexpectedSigningMethodMessage = "unexpected signing method"
	invalidTokenMessage            = "invalid token"
	refreshTokenNotFoundMessage    = "refresh token not found"

	accessTokenExpire  = time.Hour
	refreshTokenExpire = 24 * time.Hour
)

var publicPaths = map[string]struct{}{
	"/user/login":         {},
	"/user/register":      {},
	"/user/token/refresh": {},
	"/video/list":         {},
	"/video/popular":      {},
	"/video/search":       {},
}
