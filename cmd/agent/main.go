package main

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/agent"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
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

	host, reportInterval, pollInterval := mustParseArgs()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	st := storage.NewMemStorage()
	metricAgent := agent.NewMetricAgent(st, agent.NewPostSender(), reportInterval, host)
	collector := service.NewMetricCollector(st, pollInterval)

	go collector.Collect(shutdownCtx)
	go metricAgent.Push(shutdownCtx)

	<-ctx.Done()
	cancel()
	logger.Log.Info("Приложение завершено.")
}
