package token

import "strings"

// DefaultPublicPaths 未配置 Auth.PublicPaths 时使用的兜底白名单，
// 等价于配置化之前写死在中间件里的公开接口列表。
var DefaultPublicPaths = []string{
	"/users",            // POST 注册
	"/sessions",         // POST 登录
	"/sessions/refresh", // POST 刷新令牌
	"/users/*/videos",   // GET 作者视频列表（动态路径）
	"/videos/popular",   // GET 热门视频
	"/videos/search",    // GET 搜索视频
}

// Whitelist 是已编译的公开路径白名单。
// 不含通配符的模式走 map 精确匹配，含通配符的模式走前缀/中间/后缀匹配。
type Whitelist struct {
	exact    map[string]struct{}
	patterns []string
}

// NewWhitelist 把配置中的路径模式列表编译成白名单；空字符串与首尾空白会被忽略。
func NewWhitelist(patterns []string) *Whitelist {
	w := &Whitelist{
		exact:    make(map[string]struct{}, len(patterns)),
		patterns: make([]string, 0, len(patterns)),
	}
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.Contains(pattern, "*") {
			w.patterns = append(w.patterns, pattern)
			continue
		}
		w.exact[pattern] = struct{}{}
	}
	return w
}

// Contains 判断请求路径是否命中白名单。
// 接收者为 nil 时一律返回 false，避免漏配配置导致白名单被绕过（fail-closed）。
func (w *Whitelist) Contains(path string) bool {
	if w == nil {
		return false
	}
	if _, ok := w.exact[path]; ok {
		return true
	}
	for _, pattern := range w.patterns {
		if matchWildcard(pattern, path) {
			return true
		}
	}
	return false
}

// matchWildcard 按 * 通配匹配，* 可匹配任意字符（含 /）。
//
// 把 pattern 按 * 切开，每一段都是必须原样出现的字面量，匹配就是三步：
//
//	1. 首段锚定开头：path 必须以首段开头，对齐后切掉；
//	2. 末段锚定结尾：剩下的部分必须以末段结尾，对齐后切掉；
//	3. 切掉首尾后的区域完全由 * 填空，中间段只要能按顺序找到即可。
//
// 例如 /users/*/videos 命中 /users/42/videos，但不命中 /users/videos：
// 中间那个 * 至少要吃掉一个 /，末段的 /videos 还必须贴住整个 path 的结尾。
func matchWildcard(pattern, path string) bool {
	literals := strings.Split(pattern, "*")

	// 没有 * 时退化为全等比较。NewWhitelist 已把这类模式分流到 exact，
	// 这里保留是为了函数被单独调用时也安全。
	if len(literals) == 1 {
		return pattern == path
	}

	// 1. 首段锚定开头。
	body, ok := strings.CutPrefix(path, literals[0])
	if !ok {
		return false
	}

	// 2. 末段锚定结尾。cut 完首尾，body 就是交给 * 填空的中间区域。
	body, ok = strings.CutSuffix(body, literals[len(literals)-1])
	if !ok {
		return false
	}

	// 3. 中间段按顺序出现即可；每命中一段就从其后继续找，因此不会重叠。
	for _, literal := range literals[1 : len(literals)-1] {
		_, after, found := strings.Cut(body, literal)
		if !found {
			return false
		}
		body = after
	}
	return true
}
