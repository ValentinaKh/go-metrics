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
	"time"
)

func main() {
	run()
}

func run() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	st := storage.NewMemStorage()
	metricAgent := agent.NewMetricAgent(st, agent.NewPostSender(), 10*time.Second)
	collector := service.NewMetricCollector(st, 2*time.Second)

	go collector.Collect(shutdownCtx)
	go metricAgent.Push(shutdownCtx)

	<-ctx.Done()
	cancel()
	fmt.Println("Приложение завершено.")
}
