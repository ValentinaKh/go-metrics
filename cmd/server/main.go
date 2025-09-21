package main

import (
	"context"
	"database/sql"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/repository"
	"github.com/ValentinaKh/go-metrics/internal/server"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	run()
}

func run() {
	err := logger.InitializeZapLogger("info")
	if err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Приложение запускается")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	args := parseArgs()
	logger.Log.Info("Приложение работает с настройками", zap.Any("Настройки", args))

	var db *sql.DB
	if args.ConnStr != "" {
		db = repository.MustConnectDB(args.ConnStr)
	}
	server.ConfigureServer(shutdownCtx, args, db)
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	<-ctx.Done()
	cancel()

	logger.Log.Info("Приложение останавливается")
}
