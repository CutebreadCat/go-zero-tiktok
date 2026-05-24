package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockLimiter struct {
	allow   bool
	err     error
	lastKey string
}

func (m *mockLimiter) Allow(key string) (bool, error) {
	m.lastKey = key
	return m.allow, m.err
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		allow     bool
		limiterErr error
		remoteAddr string
		wantKey   string
		wantNext  bool
	}{
		{"allowed", true, nil, "1.2.3.4:8080", "1.2.3.4:8080", true},
		{"not allowed", false, nil, "1.2.3.4:8080", "1.2.3.4:8080", false},
		{"limiter error", false, errors.New("redis down"), "1.2.3.4:8080", "1.2.3.4:8080", false},
		{"empty remote addr", true, nil, "", rateLimitRemoteAddrKey, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLimiter{allow: tt.allow, err: tt.limiterErr}
			m := NewRateLimitMiddleware(ml)

			called := false
			handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			handler(httptest.NewRecorder(), r)

			if ml.lastKey != tt.wantKey {
				t.Errorf("key = %s, want %s", ml.lastKey, tt.wantKey)
			}
			if called != tt.wantNext {
				t.Errorf("next called = %v, want %v", called, tt.wantNext)
			}
		})
	}
}
