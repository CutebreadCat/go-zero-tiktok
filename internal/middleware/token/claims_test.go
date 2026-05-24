package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"nil context", nil, ""},
		{"no value", context.Background(), ""},
		{"with userID", context.WithValue(context.Background(), UserIDContextKey, "u42"), "u42"},
		{"wrong type", context.WithValue(context.Background(), UserIDContextKey, 123), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserIDFromContext(tt.ctx); got != tt.want {
				t.Errorf("UserIDFromContext = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestExtractRefreshToken(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*http.Request)
		expect string
	}{
		{
			"from cookie",
			func(r *http.Request) { r.AddCookie(&http.Cookie{Name: refreshTokenName, Value: "tok123"}) },
			"tok123",
		},
		{
			"from header",
			func(r *http.Request) { r.Header.Set(authorizationHeader, "Bearer tok456") },
			"tok456",
		},
		{
			"from form",
			func(r *http.Request) {
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.Body = http.NoBody
				r.Form = map[string][]string{refreshTokenName: {"tok789"}}
			},
			"tok789",
		},
		{
			"empty",
			func(r *http.Request) {},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			tt.setup(r)
			if got := extractRefreshToken(r); got != tt.expect {
				t.Errorf("extractRefreshToken = %s, want %s", got, tt.expect)
			}
		})
	}
}
