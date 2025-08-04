package main

import (
	"context"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/agent"
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

	host, reportInterval, pollInterval := parseFlags()

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
	fmt.Println("Приложение завершено.")
}
