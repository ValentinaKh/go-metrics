package logger

import (
	"log/slog"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

// Setup настраивает логгер и устанавливает его как глобальный. Параметр level типа string для возможности передать
// уровень логирования через флаг или env
func Setup(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	slog.SetDefault(slog.New(handler))
}

// InitializeZapLogger настраивает zap логгер
func InitializeZapLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	// Меняем формат времени
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
