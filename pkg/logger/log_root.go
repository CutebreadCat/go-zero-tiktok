package logger

import (
	"os"
	"path/filepath"
	"strings"
)

// LogRootDir 返回日志根目录，即 {service}/{date}/{service}.log 的父目录。
//
// 解析顺序：
//  1. 环境变量 TIKTOK_LOG_DIR（相对路径按当前工作目录展开）；
//  2. 从当前工作目录向上查找 go.mod，取其所在目录（项目根）下的 logs/；
//  3. 找不到项目根时，回退到当前工作目录下的 logs/。
//
// 这样无论从仓库根目录还是从服务子目录启动进程，日志都会落在同一处，
// 不会因为 cwd 不同而在 app/*/logs/ 下再生成一套。
func LogRootDir() string {
	if dir := strings.TrimSpace(os.Getenv(LogDirEnv)); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
		return dir
	}

	base := CurrentDir()
	if root, ok := findModuleRoot(base); ok {
		base = root
	}
	return filepath.Join(base, LogFilePath)
}

// findModuleRoot 从 start 目录逐级向上查找包含 go.mod 的目录。
// 到达文件系统根仍未找到时返回 false。
func findModuleRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}

	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !info.IsDir() {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
