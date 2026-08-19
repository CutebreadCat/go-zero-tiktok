package token

import "net/http"

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

// SetAccessTokenCookie 把 access token 写入 HttpOnly Cookie。
func SetAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(accessTokenExpire.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetRefreshTokenCookie 把 refresh token 写入 HttpOnly Cookie。
func SetRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(refreshTokenExpire.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetAccessTokenFromCookie 从请求 Cookie 中读取 access token。
func GetAccessTokenFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(AccessTokenCookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// GetRefreshTokenFromCookie 从请求 Cookie 中读取 refresh token。
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
