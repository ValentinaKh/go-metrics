package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/repository"
	"github.com/ValentinaKh/go-metrics/internal/server"
)

func main() {
	run()
}

func run() {
	err := logger.InitializeZapLogger("info")
	if err != nil {
		panic(err)
	}
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {
			logger.Log.Error("Ошибка при закрытии логгера", zap.Error(err))
		}
	}(logger.Log)

	logger.Log.Info("Приложение запускается")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	args := config.MustParseServerArgs()

	logger.Log.Info("Приложение работает с настройками", zap.Any("Настройки", args))

	var db *sql.DB
	if args.ConnStr != "" {
		db = repository.MustConnectDB(args.ConnStr)
	}
	server.ConfigureServer(shutdownCtx, args, db)
	defer func() {
		if db != nil {
			err := db.Close()
			if err != nil {
				return
			}
		}
	}()

	<-ctx.Done()
	cancel()

	logger.Log.Info("Приложение останавливается")
}
