package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindModuleRoot(t *testing.T) {
	root, ok := findModuleRoot(CurrentDir())
	if !ok {
		t.Fatal("expected to locate go.mod from the package directory")
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("expect go.mod under %s: %v", root, err)
	}

	// 从项目根自身出发也应得到同一个根目录。
	fromRoot, ok := findModuleRoot(root)
	if !ok || fromRoot != root {
		t.Fatalf("findModuleRoot(%q) = (%q, %v), want (%q, true)", root, fromRoot, ok, root)
	}
}

func TestLogRootDirEnvOverride(t *testing.T) {
	dir := filepath.Clean(t.TempDir())
	t.Setenv(LogDirEnv, dir)

	if got := LogRootDir(); got != dir {
		t.Fatalf("LogRootDir() = %q, want %q", got, dir)
	}
}

func TestLogRootDirAnchoredToProjectRoot(t *testing.T) {
	t.Setenv(LogDirEnv, "")

	root, ok := findModuleRoot(CurrentDir())
	if !ok {
		t.Skip("go.mod not found, cannot assert project-root anchoring")
	}

	want := filepath.Join(root, LogFilePath)
	if got := LogRootDir(); got != want {
		t.Fatalf("LogRootDir() = %q, want %q", got, want)
	}
}

// TestLogRootDirIndependentOfCwd 回归用例：从服务子目录启动进程时，
// 日志根目录必须仍然指向项目根，而不是在服务目录下再生成一套 logs/。
func TestLogRootDirIndependentOfCwd(t *testing.T) {
	root, ok := findModuleRoot(CurrentDir())
	if !ok {
		t.Skip("go.mod not found, cannot assert project-root anchoring")
	}

	t.Setenv(LogDirEnv, "")
	want := filepath.Join(root, LogFilePath)

	sub := filepath.Join(root, "app", "video", "rpc")
	if info, err := os.Stat(sub); err != nil || !info.IsDir() {
		t.Skipf("service directory %s not available", sub)
	}
	t.Chdir(sub)

	if got := LogRootDir(); got != want {
		t.Fatalf("from %s: LogRootDir() = %q, want %q", sub, got, want)
	}
}

// TestUpdateLoggerWritesUnderLogRoot 验证日志不再写进进程 cwd，而是落在日志根目录下。
func TestUpdateLoggerWritesUnderLogRoot(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	t.Setenv(LogDirEnv, root)

	if err := updateLogger("test-svc"); err != nil {
		t.Fatalf("updateLogger: %v", err)
	}
	t.Cleanup(func() { _ = Close() })

	Info("hello from log root test")
	if err := Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	date := time.Now().Format("2006-01-02")
	for _, name := range []string{"test-svc.log", "test-svc_stderr.log"} {
		path := filepath.Join(root, "test-svc", date, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expect %s: %v", path, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, "test-svc", date, "test-svc.log"))
	if err != nil {
		t.Fatalf("read main log: %v", err)
	}
	if !strings.Contains(string(data), "hello from log root test") {
		t.Fatalf("main log does not contain the record: %s", data)
	}
}
