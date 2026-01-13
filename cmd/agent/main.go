package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/agent"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	run()
}

func run() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date:  %s\n", buildDate)
	fmt.Printf("Build commit:  %s\n", buildCommit)

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

	logger.Log.Info("Приложение запущено.")

	args := config.MustParseAgentArgs()
	retryConfig := &config.RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
	err = retryConfig.Validate()
	if err != nil {
		panic(err)
	}
	logger.Log.Info("Приложение работает с настройками", zap.Any("Настройки", args))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	mChan := make(chan []models.Metrics, 10)

	wgGroup, err := agent.ConfigureAgent(shutdownCtx, args, retryConfig, mChan)
	if err != nil {
		logger.Log.Fatal("Ошибка при конфигурации агента", zap.Error(err))
	}

	<-ctx.Done()
	cancel()
	wgGroup.Wait()

	//закрываем канал только после того, как все продюсеры остановились
	close(mChan)
	logger.Log.Info("Приложение завершено.")
}
