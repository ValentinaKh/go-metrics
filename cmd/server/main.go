package main

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/handler/middleware"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/ValentinaKh/go-metrics/internal/storage/decorator"
	"github.com/go-chi/chi/v5"
	"net/http"
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

	logger.Log.Info("Приложение запускается")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownCtx, cancel := context.WithCancel(context.Background())

	host, interval, fileName, restore := parseArgs()

	memSt := storage.NewMemStorage()

	if restore {
		err := service.LoadMetrics(fileName, memSt)
		if err != nil {
			panic(err)
		}
	}
	metricStorage := createMetricStorage(shutdownCtx, interval, fileName, memSt)
	createServer(service.NewMetricsService(metricStorage), host)

	<-ctx.Done()
	cancel()
	shutdown(metricStorage)
	logger.Log.Info("Приложение останавливается")
}

func createServer(metricsService *service.MetricsService, host string) {
	r := chi.NewRouter()
	r.With(middleware.LoggingMw, middleware.GzipMW).Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllMetricsHandler(metricsService))
		r.With(middleware.ValidationURLRqMw).Post("/update/{type}/{name}/{value}", handler.MetricsHandler(metricsService))
		r.Post("/update/", handler.JSONUpdateMetricsHandler(metricsService))
		r.Get("/value/{type}/{name}", handler.GetMetricHandler(metricsService))
		r.Post("/value/", handler.GetJSONMetricHandler(metricsService))
	})

	srv := &http.Server{
		Addr:    host,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func createMetricStorage(ctx context.Context, interval time.Duration, fileName string, memStorage *storage.MemStorage) service.Storage {
	if interval == 0 {
		st, err := decorator.NewStoreWithSyncFile(memStorage, fileName)
		if err != nil {
			panic(err)
		}
		return st
	} else {
		st, err := decorator.NewStoreWithAsyncFile(ctx, memStorage, interval, fileName)
		if err != nil {
			panic(err)
		}
		return st
	}
}

func shutdown(st service.Storage) {
	if s, ok := st.(*decorator.StoreWithSyncFile); ok {
		err := s.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}
