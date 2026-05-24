package token

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsPublicPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/user/login", true},
		{"/user/register", true},
		{"/video/popular", true},
		{"/user/profile", false},
		{"/video/publish", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isPublicPath(tt.path); got != tt.want {
				t.Errorf("isPublicPath(%s) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestExtractAccessToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer", "Bearer tok123", "tok123"},
		{"case insensitive bearer", "bearer tok123", "tok123"},
		{"missing header", "", ""},
		{"invalid format", "Basic tok123", ""},
		{"no token value", "Bearer", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set(authorizationHeader, tt.header)
			}
			if got := extractAccessToken(r); got != tt.want {
				t.Errorf("extractAccessToken = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAuthMiddleware_PublicPath(t *testing.T) {
	m := AuthMiddleware(testSecret)
	called := false
	handler := m(func(w http.ResponseWriter, r *http.Request) { called = true })

	r := httptest.NewRequest(http.MethodGet, "/user/login", nil)
	handler(httptest.NewRecorder(), r)

	if !called {
		t.Error("public path should pass through")
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	m := AuthMiddleware(testSecret)
	called := false
	handler := m(func(w http.ResponseWriter, r *http.Request) { called = true })

	r := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	handler(httptest.NewRecorder(), r)

	if called {
		t.Error("should reject request without token")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	m := AuthMiddleware(testSecret)
	called := false
	handler := m(func(w http.ResponseWriter, r *http.Request) { called = true })

	r := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	r.Header.Set(authorizationHeader, "Bearer invalid.token.here")
	handler(httptest.NewRecorder(), r)

	if called {
		t.Error("should reject invalid token")
	}
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	token, _ := GenerateRefreshToken(testSecret, "u1")
	m := AuthMiddleware(testSecret)
	called := false
	handler := m(func(w http.ResponseWriter, r *http.Request) { called = true })

	r := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	r.Header.Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	handler(httptest.NewRecorder(), r)

	if called {
		t.Error("should reject refresh token for access middleware")
	}
}

func TestAuthMiddleware_ValidToken_SetsUserID(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, "u42")
	m := AuthMiddleware(testSecret)

	var gotUserID string
	handler := m(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
	})

	r := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	r.Header.Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	handler(httptest.NewRecorder(), r)

	if gotUserID != "u42" {
		t.Errorf("UserID in context = %s, want u42", gotUserID)
	}
}
