package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// log writes through the logger currently installed in control. Keeping the
// lookup here makes all package-level output methods safe while allowing the
// active logger to be replaced at runtime.
func (c *controlLogger) log(level zapcore.Level, msg string, fields ...zap.Field) {
	if l := c.snapshot(); l != nil {
		if entry := l.Check(level, msg); entry != nil {
			entry.Write(fields...)
		}
	}
}

func (c *controlLogger) Debug(msg string, fields ...zap.Field) { c.log(zap.DebugLevel, msg, fields...) }
func (c *controlLogger) Info(msg string, fields ...zap.Field)  { c.log(zap.InfoLevel, msg, fields...) }
func (c *controlLogger) Warn(msg string, fields ...zap.Field)  { c.log(zap.WarnLevel, msg, fields...) }
func (c *controlLogger) Error(msg string, fields ...zap.Field) { c.log(zap.ErrorLevel, msg, fields...) }
func (c *controlLogger) DPanic(msg string, fields ...zap.Field) {
	c.log(zap.DPanicLevel, msg, fields...)
}
func (c *controlLogger) Panic(msg string, fields ...zap.Field) { c.log(zap.PanicLevel, msg, fields...) }
func (c *controlLogger) Fatal(msg string, fields ...zap.Field) { c.log(zap.FatalLevel, msg, fields...) }

func Debug(msg string, fields ...zap.Field)  { output(zap.DebugLevel, msg, fields...) }
func Info(msg string, fields ...zap.Field)   { output(zap.InfoLevel, msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { output(zap.WarnLevel, msg, fields...) }
func Error(msg string, fields ...zap.Field)  { output(zap.ErrorLevel, msg, fields...) }
func DPanic(msg string, fields ...zap.Field) { output(zap.DPanicLevel, msg, fields...) }
func Panic(msg string, fields ...zap.Field)  { output(zap.PanicLevel, msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { output(zap.FatalLevel, msg, fields...) }

func Debugf(format string, args ...interface{}) { Debug(fmt.Sprintf(format, args...)) }
func Infof(format string, args ...interface{})  { Info(fmt.Sprintf(format, args...)) }
func Warnf(format string, args ...interface{})  { Warn(fmt.Sprintf(format, args...)) }
func Errorf(format string, args ...interface{}) { Error(fmt.Sprintf(format, args...)) }

// output locks only while reading the active logger pointer. The lock is
// released before zap performs level checks, encoding, or writing.
func output(level zapcore.Level, msg string, fields ...zap.Field) {
	control.loggerMu.RLock()
	l := control.logger
	control.loggerMu.RUnlock()
	if l == nil {
		return
	}
	if entry := l.Check(level, msg); entry != nil {
		entry.Write(fields...)
	}
}
