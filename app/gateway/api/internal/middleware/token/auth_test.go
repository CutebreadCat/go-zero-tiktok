package token

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtpkg "go_zero-tiktok/pkg/jwt"

	"github.com/golang-jwt/jwt/v4"
)

const testSecret = "unit-test-secret"

// invokeMiddleware 使用默认白名单调用鉴权中间件，
// 返回 next 是否被放行，以及下游从上下文读到的 user_id。
func invokeMiddleware(t *testing.T, path, authorization string) (called bool, userID int64) {
	t.Helper()
	return invokeMiddlewareWithWhitelist(t, path, authorization, NewWhitelist(DefaultPublicPaths))
}

// invokeMiddlewareWithWhitelist 同上，但使用调用方指定的白名单，用于验证配置驱动的行为。
func invokeMiddlewareWithWhitelist(t *testing.T, path, authorization string, whitelist *Whitelist) (called bool, userID int64) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()

	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		userID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}

	AuthMiddleware(testSecret, whitelist)(next)(rec, req)
	return called, userID
}

func mustAccessToken(t *testing.T, userID string) string {
	t.Helper()

	token, err := jwtpkg.GenerateAccessToken(testSecret, userID)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	return token
}

// mustAccessTokenWithoutUserID 构造一个签名合法、但 payload 中不含 user_id 声明的 access token。
func mustAccessTokenWithoutUserID(t *testing.T) string {
	t.Helper()

	now := time.Now()
	claims := jwt.MapClaims{
		"token_type": jwtpkg.AccessTokenType,
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(time.Hour).Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func TestAuthMiddleware_ValidUserIDPasses(t *testing.T) {
	called, userID := invokeMiddleware(t, "/videos/publish", "Bearer "+mustAccessToken(t, "123"))
	if !called {
		t.Fatal("expected request to pass, but it was rejected")
	}
	if userID != 123 {
		t.Fatalf("user_id in context = %d, want 123", userID)
	}
}

func TestAuthMiddleware_InvalidUserIDRejected(t *testing.T) {
	cases := []struct {
		name  string
		token string
	}{
		{name: "non numeric", token: mustAccessToken(t, "abc")},
		{name: "zero", token: mustAccessToken(t, "0")},
		{name: "negative", token: mustAccessToken(t, "-5")},
		{name: "empty", token: mustAccessToken(t, "")},
		{name: "missing claim", token: mustAccessTokenWithoutUserID(t)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called, _ := invokeMiddleware(t, "/videos/publish", "Bearer "+tc.token)
			if called {
				t.Fatal("expected request to be rejected, but it passed")
			}
		})
	}
}

func TestAuthMiddleware_InvalidSignatureRejected(t *testing.T) {
	foreign, err := jwtpkg.GenerateAccessToken("another-secret", "123")
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	called, _ := invokeMiddleware(t, "/videos/publish", "Bearer "+foreign)
	if called {
		t.Fatal("expected request with invalid signature to be rejected")
	}
}

func TestAuthMiddleware_ProtectedPathWithoutTokenRejected(t *testing.T) {
	called, _ := invokeMiddleware(t, "/videos/publish", "")
	if called {
		t.Fatal("expected request without token to be rejected")
	}
}

func TestAuthMiddleware_PublicPathAllowsAnonymous(t *testing.T) {
	called, userID := invokeMiddleware(t, "/videos/popular", "")
	if !called {
		t.Fatal("expected public path to pass without token")
	}
	if userID != 0 {
		t.Fatalf("anonymous user_id = %d, want 0", userID)
	}
}

func TestAuthMiddleware_CustomWhitelistOverridesDefault(t *testing.T) {
	custom := NewWhitelist([]string{"/custom/public"})

	called, _ := invokeMiddlewareWithWhitelist(t, "/custom/public", "", custom)
	if !called {
		t.Fatal("expected custom public path to pass without token")
	}

	called, _ = invokeMiddlewareWithWhitelist(t, "/videos/popular", "", custom)
	if called {
		t.Fatal("expected default public path to be protected once whitelist is overridden")
	}
}
