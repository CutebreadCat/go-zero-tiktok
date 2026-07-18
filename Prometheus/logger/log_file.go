package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// openLogFiles creates the daily service log files.
func openLogFiles(service string) (io.WriteCloser, io.WriteCloser, error) {
	dir := filepath.Join(CurrentDir(), LogFilePath, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	mainFile, err := newRollingFile(filepath.Join(dir, service+".log"))
	if err != nil {
		return nil, nil, err
	}
	errFile, err := newRollingFile(filepath.Join(dir, service+"_stderr.log"))
	if err != nil {
		_ = mainFile.Close()
		return nil, nil, err
	}
	return mainFile, errFile, nil
}

// buildLogCores routes normal records to the main file and errors to stderr.
func buildLogCores(mainFile, errFile io.Writer, level zapcore.Level) []zapcore.Core {
	mainEnabled := zap.LevelEnablerFunc(func(entryLevel zapcore.Level) bool {
		return entryLevel >= level && entryLevel < zapcore.ErrorLevel
	})
	errEnabled := zap.LevelEnablerFunc(func(entryLevel zapcore.Level) bool { return entryLevel >= zapcore.ErrorLevel })
	return []zapcore.Core{
		zapcore.NewCore(defaultEnc(), zapcore.Lock(zapcore.AddSync(mainFile)), mainEnabled),
		zapcore.NewCore(defaultEnc(), zapcore.Lock(zapcore.AddSync(errFile)), errEnabled),
		zapcore.NewCore(defaultEnc(), zapcore.Lock(zapcore.AddSync(os.Stdout)), mainEnabled),
	}
}

// rollingFile limits one file to 10 MB and keeps seven rotated backups.
type rollingFile struct {
	mu         sync.Mutex
	file       *os.File
	path       string
	maxSize    int64
	maxBackups int
}

func newRollingFile(path string) (*rollingFile, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &rollingFile{file: f, path: path, maxSize: 10 << 20, maxBackups: 7}, nil
}

func (r *rollingFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if info, err := r.file.Stat(); err == nil && info.Size()+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	return r.file.Write(p)
}

func (r *rollingFile) rotate() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	for i := r.maxBackups - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", r.path, i)
		next := fmt.Sprintf("%s.%d", r.path, i+1)
		if _, err := os.Stat(old); err == nil {
			_ = os.Rename(old, next)
		}
	}
	_ = os.Rename(r.path, r.path+".1")
	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	r.file = f
	return nil
}

func (r *rollingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}
