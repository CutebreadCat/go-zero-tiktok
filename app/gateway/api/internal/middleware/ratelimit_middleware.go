// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"net/http"
	"sync"
	"time"
)

const (
	// rateLimitWindow 限流窗口
	rateLimitWindow = time.Second
	// rateLimitMaxPerWindow 单 IP 每个窗口内的最大请求数
	rateLimitMaxPerWindow = 100
)

type RateLimitMiddleware struct {
	mu     sync.Mutex
	window time.Time
	counts map[string]int
}

func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{counts: make(map[string]int)}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if key == "" {
			key = "unknown"
		}
		m.mu.Lock()
		now := time.Now()
		if m.window.IsZero() || now.Sub(m.window) >= rateLimitWindow {
			m.window = now
			m.counts = make(map[string]int)
		}
		m.counts[key]++
		allowed := m.counts[key] <= rateLimitMaxPerWindow
		m.mu.Unlock()
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
