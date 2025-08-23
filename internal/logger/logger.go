package logger

import (
	"log/slog"
	"os"
)

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
