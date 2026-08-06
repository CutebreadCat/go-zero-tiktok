package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// parseLevel converts an English log level name to zap's level type.
// Names are case-insensitive; unknown values fall back to InfoLevel.
func parseLevel(value string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return zapcore.DebugLevel
	case "info", "information":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// ParseLevel is the exported form for configuration packages outside logger.
func ParseLevel(value string) zapcore.Level { return parseLevel(value) }

type config struct {
	core zapcore.Core
	enc  zapcore.Encoder
	ws   zapcore.WriteSyncer
	lvl  zapcore.Level
}

func buildConfig(core zapcore.Core) *config {
	cfg := defaultConfig()
	cfg.core = core
	if core == nil {
		cfg.core = zapcore.NewCore(cfg.enc, cfg.ws, cfg.lvl)
	}
	return cfg
}

func BuildConfig(cfg *config, opts ...zap.Option) *zap.Logger {
	if cfg == nil {
		cfg = defaultConfig()
	}
	return zap.New(cfg.core, opts...)
}

func defaultConfig() *config {
	return &config{

		enc: defaultEnc(),
		ws:  defaultWS(),
		lvl: defaultLvl(),
	}

}

func defaultEnc() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(cfg)
}
func defaultWS() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}
func defaultLvl() zapcore.Level {
	return zapcore.InfoLevel
}
