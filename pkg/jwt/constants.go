package token

import "time"

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
	RefreshPrefix    = "refresh_token:"

	unexpectedSigningMethodMessage = "unexpected signing method"
	invalidTokenMessage            = "invalid token"
	refreshTokenNotFoundMessage    = "refresh token not found"

	accessTokenExpire  = time.Hour
	refreshTokenExpire = 24 * time.Hour
)
