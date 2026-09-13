package token

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	contract "go_zero-tiktok/pkg/contract"
	jwtpkg "go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.uber.org/zap"
)

// WithAuth 装配鉴权中间件；publicPaths 为空时回退到 DefaultPublicPaths。
// 白名单在启动阶段编译一次，请求链路上只做查询。
func WithAuth(secret string, publicPaths []string) rest.RunOption {
	if len(publicPaths) == 0 {
		publicPaths = DefaultPublicPaths
	}
	whitelist := NewWhitelist(publicPaths)

	return func(server *rest.Server) {
		rest.WithUnauthorizedCallback(UnauthorizedCallback)(server)
		server.Use(AuthMiddleware(secret, whitelist))
	}
}

func UnauthorizedCallback(w http.ResponseWriter, r *http.Request, err error) {
	var codeErr *xerr.CodeError
	// 判断的是实际传入的 err，而非新建的错误，避免恒为 true
	if !errors.As(err, &codeErr) {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, codeErr.HandleResponse())
}

func AuthMiddleware(secret string, whitelist *Whitelist) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if whitelist.Contains(r.URL.Path) {
				next(w, r)
				return
			}

			accessToken := extractAccessToken(r)
			if accessToken == "" {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			claims, err := jwtpkg.ParseToken(secret, accessToken)
			if err != nil || claims.TokenType != jwtpkg.AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			// 身份提取必须 fail-closed：user_id 缺失、无法解析为非负整数或 <= 0 时一律拒绝，
			// 避免以 0 号幽灵身份放行、污染下游数据。
			userID, err := strconv.ParseInt(claims.UserID, 10, 64)
			if err != nil || userID <= 0 {
				logger.WithContext(r.Context()).Warn("auth: invalid user_id in access token",
					zap.String("path", r.URL.Path),
					zap.String("reason", invalidUserIDReason(claims.UserID, err)),
				)
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			ctx := context.WithValue(r.Context(), contract.ContextKeyUserID, userID)
			next(w, r.WithContext(ctx))
		}
	}
}

// invalidUserIDReason 归纳身份提取失败的简要原因，仅用于告警日志（不包含令牌内容）。
func invalidUserIDReason(raw string, err error) string {
	switch {
	case raw == "":
		return "user_id missing"
	case err != nil:
		return "user_id not numeric"
	default:
		return "user_id not positive"
	}
}

// extractAccessToken 优先从 Cookie 读取 access_token，取不到再尝试 Authorization: Bearer。
func extractAccessToken(r *http.Request) string {
	if token, err := jwtpkg.GetAccessTokenFromCookie(r); err == nil && token != "" {
		return token
	}

	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func UserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	userID, _ := ctx.Value(contract.ContextKeyUserID).(int64)
	return userID
}
