package main

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/agent"
	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
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

	host, reportInterval, pollInterval := mustParseArgs()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	st := storage.NewMemStorage()
	retryConfig := config.RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
	err = retryConfig.Validate()
	if err != nil {
		panic(err)
	}
	metricAgent := agent.NewMetricAgent(st, agent.NewPostSender(host,
		retry.NewRetrier(
			retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), retryConfig.MaxAttempts),
			retry.NewStaticDelayStrategy(retryConfig.Delays),
			&retry.SleepTimeProvider{})), reportInterval)
	collector := service.NewMetricCollector(st, pollInterval)

	go collector.Collect(shutdownCtx)
	go metricAgent.Push(shutdownCtx)

	<-ctx.Done()
	cancel()
	logger.Log.Info("Приложение завершено.")
}
