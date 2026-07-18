package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type controlLogger struct {
	// loggerMu protects access to the active logger pointer only. It is never
	// held while zap encodes or writes a log entry.
	loggerMu      sync.RWMutex
	logger        *Logger
	hooksMu       sync.RWMutex
	hooks         []func(zapcore.Entry) error
	contextMu     sync.RWMutex
	contextFields []func(context.Context) []zap.Field
	done          atomic.Bool
	rotationMu    sync.Mutex
	rotationStop  chan struct{}
}

// setLogger replaces the active logger used by the package-level facade.
func (c *controlLogger) setLogger(l *Logger) {
	c.loggerMu.Lock()
	old := c.logger
	c.logger = l
	c.loggerMu.Unlock()
	if old != nil && old != l {
		_ = old.Close()
	}
}

// snapshot returns the current logger and releases the lock before callers
// perform any logger operation.
func (c *controlLogger) snapshot() *Logger {
	c.loggerMu.RLock()
	l := c.logger
	c.loggerMu.RUnlock()
	return l
}

type Logger struct {
	*zap.Logger
	closers []io.Closer
}

var (
	control                controlLogger
	logLevel               = zapcore.InfoLevel
	callerSkip             = 2
	defaultSDervice        = "_default"
	errorSpanLevel         = zapcore.ErrorLevel
	recodeStackTraceInSpan = false
)

// SetLogger installs the logger used by the package-level logging functions.
func SetLogger(l *zap.Logger) {
	if l == nil {
		control.setLogger(nil)
		return
	}
	control.setLogger(&Logger{Logger: l})
}

// Logger returns the currently configured logger, or nil when logging has not
// been initialized yet.
func Current() *Logger { return control.snapshot() }

// AddLoggerHook 注册日志写出后的回调函数。
func AddLoggerHook(fns ...func(zapcore.Entry) error) {
	control.hooksMu.Lock()
	control.hooks = append(control.hooks, fns...)
	control.hooksMu.Unlock()
}

func Sync() error {
	if l := control.snapshot(); l != nil {
		return l.Sync()
	}
	return nil
}

func Close() error {
	control.rotationMu.Lock()
	if control.rotationStop != nil {
		close(control.rotationStop)
		control.rotationStop = nil
	}
	control.done.Store(false)
	control.rotationMu.Unlock()
	control.loggerMu.Lock()
	l := control.logger
	control.logger = nil
	control.loggerMu.Unlock()
	if l == nil {
		return nil
	}
	return l.Close()
}

func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	err := l.Sync()
	if len(l.closers) > 0 {
		for _, closer := range l.closers {
			if closeErr := closer.Close(); err == nil {
				err = closeErr
			}
		}
	}
	return err
}

// init installs a usable file logger as soon as this package is loaded. This
// keeps package-level logging safe before main has selected its final path.
func init() {
	control.hooks = make([]func(zapcore.Entry) error, 0)
	control.contextFields = []func(context.Context) []zap.Field{defaultContextFields}
	InitStdout()
}

// CurrentDir returns the process working directory. It falls back to the
// current directory when the operating system cannot resolve it.
func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// UpdateLogger replaces the active logger and writes to the service-specific
// files under the current working directory.
func UpdateLogger(service string) error { return updateLogger(service) }

func updateLogger(service string) error {
	if service == "" {
		service = defaultSDervice
	}
	mainFile, errFile, err := openLogFiles(service)
	if err != nil {
		return err
	}
	l := &Logger{
		Logger:  zap.New(zapcore.NewTee(buildLogCores(mainFile, errFile, logLevel)...), loggerOptions(service)...),
		closers: []io.Closer{mainFile, errFile},
	}
	control.setLogger(l)
	return nil
}

func loggerOptions(service string) []zap.Option {
	control.hooksMu.RLock()
	hooks := append([]func(zapcore.Entry) error(nil), control.hooks...)
	control.hooksMu.RUnlock()
	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(callerSkip), zap.Fields(
		zap.String(ServiceKey, service),
		zap.String(SourceKey, fmt.Sprintf("app-%s", service)),
	)}
	if len(hooks) > 0 {
		opts = append(opts, zap.Hooks(hooks...))
	}
	return opts
}

// Init 按服务名和日志级别初始化日志，并启动每日自动轮换。
func Init(service string, level string) {
	if service == "" {
		panic("service should not be empty")
	}
	logLevel = parseLevel(level)
	if err := updateLogger(service); err != nil {
		panic(err)
	}
	control.scheduleUpdateLogger(service)
}

// scheduleUpdateLogger 每天零点重新打开当天的日志文件。
func (c *controlLogger) scheduleUpdateLogger(service string) {
	if c.done.Swap(true) {
		return
	}
	c.rotationMu.Lock()
	c.rotationStop = make(chan struct{})
	stop := c.rotationStop
	c.rotationMu.Unlock()
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
			timer := time.NewTimer(time.Until(next))
			select {
			case <-timer.C:
				_ = updateLogger(service)
			case <-stop:
				timer.Stop()
				return
			}
		}
	}()
}

// Init initializes or replaces the package logger with a file at path. The
// parent directory is created when necessary.
func InitWithPath(path string, opts ...zap.Option) (*Logger, error) {
	if path == "" {
		return nil, os.ErrInvalid
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	cfg.ws = zapcore.AddSync(file)
	cfg.core = zapcore.NewCore(cfg.enc, cfg.ws, cfg.lvl)
	l := &Logger{Logger: BuildConfig(cfg, opts...), closers: []io.Closer{file}}
	control.setLogger(l)
	return l, nil
}

// InitWithPath is kept as an explicit alias for callers that prefer the
// descriptive name.
// InitStdout replaces the active logger with a standard-output logger.
func InitStdout(opts ...zap.Option) *Logger {
	cfg := buildConfig(nil)
	l := &Logger{Logger: BuildConfig(cfg, opts...)}
	control.setLogger(l)
	return l
}
