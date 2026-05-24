// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"go_zero-tiktok/internal/middleware/government/limiter"
	"go_zero-tiktok/internal/shared/xerr"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type RateLimitMiddleware struct {
	limiter limiter.Limiter
}

func NewRateLimitMiddleware(l limiter.Limiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: l}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation
		allowed, err := m.limiter.Allow(rateLimitKey(r))

		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.NewServerBusy())
			return
		}

		if !allowed {
			httpx.ErrorCtx(r.Context(), w, xerr.NewServerBusy())
			return
		}

		next(w, r)
	}
}

// Passthrough to next handler if need
func rateLimitKey(r *http.Request) string {
	if r.RemoteAddr == "" {
		return rateLimitRemoteAddrKey
	}
	return r.RemoteAddr
}
