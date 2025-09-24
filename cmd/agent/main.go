package main

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/agent"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	mChan := make(chan []models.Metrics, 10)

	wgGroup := agent.ConfigureAgent(shutdownCtx, args, retryConfig, mChan)

	<-ctx.Done()
	cancel()
	wgGroup.Wait()

	//закрываем канал только после того, как все продюсеры остановились
	close(mChan)
	logger.Log.Info("Приложение завершено.")
}
