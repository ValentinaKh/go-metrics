package main

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/agent"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	run()
}

func run() {
	logger.Setup("info")

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
	slog.Info("Приложение завершено.")
}
