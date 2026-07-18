package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go_zero-tiktok/pkg/ctxkey"
)

func defaultContextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 3)
	for _, item := range []struct {
		key  ctxkey.Key
		name string
	}{
		{ctxkey.UserID, "user_id"},
		{ctxkey.TraceID, "trace_id"},
		{ctxkey.SpanID, "span_id"},
	} {
		if value, ok := ctx.Value(item.key).(string); ok && value != "" {
			fields = append(fields, zap.String(item.name, value))
		}
	}
	return fields
}

// ContextLogger carries request context and structured fields through a
// chainable logging API. It is immutable: With returns a new value, so one
// logger can safely be reused by concurrent requests.
type ContextLogger struct {
	ctx    context.Context
	fields []zap.Field
}

// WithContext creates a logger bound to ctx.
func WithContext(ctx context.Context) *ContextLogger {
	if ctx == nil {
		ctx = context.Background()
	}
	return &ContextLogger{ctx: ctx}
}

// With returns a logger with additional fields.
func (l *ContextLogger) With(fields ...zap.Field) *ContextLogger {
	if l == nil {
		return WithContext(context.Background()).With(fields...)
	}
	merged := make([]zap.Field, 0, len(l.fields)+len(fields))
	merged = append(merged, l.fields...)
	merged = append(merged, fields...)
	return &ContextLogger{ctx: l.ctx, fields: merged}
}

func (l *ContextLogger) fieldsForEntry() []zap.Field {
	if l == nil {
		return nil
	}
	fields := make([]zap.Field, 0, len(l.fields)+2)
	fields = append(fields, l.fields...)
	control.contextMu.RLock()
	extractors := append([]func(context.Context) []zap.Field(nil), control.contextFields...)
	control.contextMu.RUnlock()
	for _, extractor := range extractors {
		if extractor != nil {
			fields = append(fields, extractor(l.ctx)...)
		}
	}
	return fields
}

// AddContextFieldExtractor registers a context-to-fields adapter. This is the
// extension point for trace_id, span_id, tenant_id and similar metadata.
func AddContextFieldExtractor(extractor func(context.Context) []zap.Field) {
	if extractor == nil {
		return
	}
	control.contextMu.Lock()
	control.contextFields = append(control.contextFields, extractor)
	control.contextMu.Unlock()
}

func (l *ContextLogger) Debug(msg string, fields ...zap.Field) {
	l.write(zap.DebugLevel, msg, fields...)
}
func (l *ContextLogger) Info(msg string, fields ...zap.Field) { l.write(zap.InfoLevel, msg, fields...) }
func (l *ContextLogger) Warn(msg string, fields ...zap.Field) { l.write(zap.WarnLevel, msg, fields...) }
func (l *ContextLogger) Error(msg string, fields ...zap.Field) {
	l.write(zap.ErrorLevel, msg, fields...)
}
func (l *ContextLogger) DPanic(msg string, fields ...zap.Field) {
	l.write(zap.DPanicLevel, msg, fields...)
}
func (l *ContextLogger) Panic(msg string, fields ...zap.Field) {
	l.write(zap.PanicLevel, msg, fields...)
}
func (l *ContextLogger) Fatal(msg string, fields ...zap.Field) {
	l.write(zap.FatalLevel, msg, fields...)
}

func (l *ContextLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}
func (l *ContextLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}
func (l *ContextLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}
func (l *ContextLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *ContextLogger) write(level zapcore.Level, msg string, fields ...zap.Field) {
	fields = append(l.fieldsForEntry(), fields...)
	output(level, msg, fields...)
}
