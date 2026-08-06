package token

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	contract "go_zero-tiktok/pkg/contract"
	jwtpkg "go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func WithAuth(secret string) rest.RunOption {
	return func(server *rest.Server) {
		rest.WithUnauthorizedCallback(UnauthorizedCallback)(server)
		server.Use(AuthMiddleware(secret))
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

func AuthMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next(w, r)
				return
			}

			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			claims, err := jwtpkg.ParseToken(secret, parts[1])
			if err != nil || claims.TokenType != jwtpkg.AccessTokenType {
				httpx.ErrorCtx(r.Context(), w, xerr.NewUnauthorized("token invalid"))
				return
			}

			userID, _ := strconv.ParseInt(claims.UserID, 10, 64)
			ctx := context.WithValue(r.Context(), contract.ContextKeyUserID, userID)
			next(w, r.WithContext(ctx))
		}
	}
}

var publicPaths = map[string]struct{}{
	// 账户公开口
	"/users":            {}, // POST 注册
	"/sessions":         {}, // POST 登录
	"/sessions/refresh": {}, // POST 刷新令牌
	// 视频公开口（以 * 结尾表示前缀匹配，如动态路径 /users/:id/videos）
	"/users/*/videos": {}, // GET 作者视频列表
	"/videos/popular": {}, // GET 热门视频
	"/videos/search":  {}, // GET 搜索视频
}

// isPublicPath 判断请求路径是否命中公开白名单。
// 支持以 * 结尾的前缀匹配，用于带动态路径参数的公开接口。
func isPublicPath(path string) bool {
	if _, ok := publicPaths[path]; ok {
		return true
	}
	for p := range publicPaths {
		if strings.HasSuffix(p, "*") && strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
			return true
		}
	}
	return false
}

func UserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	userID, _ := ctx.Value(contract.ContextKeyUserID).(int64)
	return userID
}
