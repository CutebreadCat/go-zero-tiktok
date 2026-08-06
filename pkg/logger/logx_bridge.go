package logger

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logxBridge implements logx.Writer and forwards go-zero framework logs
// (HTTP access, gRPC errors, etcd, ...) into the app logger, so that they
// land in logs/{service}/{date}/ alongside business logs.
type logxBridge struct{}

// RegisterLogxBridge installs a bridge that redirects go-zero logx output
// into the app logger. Call it AFTER rest.MustNewServer / zrpc.MustNewServer,
// because those call logx.SetUp internally which would otherwise reset the
// writer back to the console/file writer.
func RegisterLogxBridge() {
	logx.SetWriter(&logxBridge{})
}

func (b *logxBridge) Alert(v any) {
	b.write(zapcore.ErrorLevel, v)
}

func (b *logxBridge) Close() error {
	// the app logger owns its file lifecycle, do not close it here
	return nil
}

func (b *logxBridge) Debug(v any, fields ...logx.LogField) {
	b.write(zapcore.DebugLevel, v, fields...)
}

func (b *logxBridge) Error(v any, fields ...logx.LogField) {
	b.write(zapcore.ErrorLevel, v, fields...)
}

func (b *logxBridge) Info(v any, fields ...logx.LogField) {
	b.write(zapcore.InfoLevel, v, fields...)
}

func (b *logxBridge) Severe(v any) {
	b.write(zapcore.ErrorLevel, v)
}

func (b *logxBridge) Slow(v any, fields ...logx.LogField) {
	b.write(zapcore.WarnLevel, v, fields...)
}

func (b *logxBridge) Stack(v any) {
	b.write(zapcore.ErrorLevel, v)
}

func (b *logxBridge) Stat(v any, fields ...logx.LogField) {
	b.write(zapcore.InfoLevel, v, fields...)
}

func (b *logxBridge) write(level zapcore.Level, v any, fields ...logx.LogField) {
	control.loggerMu.RLock()
	l := control.logger
	control.loggerMu.RUnlock()
	if l == nil {
		return
	}
	// disable the auto caller: logx already resolved the real caller into
	// fields, keeping the auto one would point at this bridge instead.
	zl := l.Logger.WithOptions(zap.WithCaller(false))
	if entry := zl.Check(level, fmt.Sprint(v)); entry != nil {
		entry.Write(convertLogxFields(fields...)...)
	}
}

// convertLogxFields maps logx fields into zap fields, normalizing trace/span
// keys to trace_id/span_id so they match the OTel extractor output.
func convertLogxFields(fields ...logx.LogField) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch f.Key {
		case "trace":
			out = append(out, zap.String("trace_id", fmt.Sprint(f.Value)))
		case "span":
			out = append(out, zap.String("span_id", fmt.Sprint(f.Value)))
		default:
			if f.Key != "" {
				out = append(out, zap.Any(f.Key, f.Value))
			}
		}
	}
	return out
}
