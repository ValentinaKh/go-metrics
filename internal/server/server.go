package server

import (
	"context"
	"database/sql"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/fileworker"
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/handler/middleware"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/repository"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/ValentinaKh/go-metrics/internal/storage/decorator"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func ConfigureServer(ctx context.Context, cfg *config.ServerArg, db *sql.DB) {
	var strg service.Storage
	var healthService handler.HealthChecker

	if cfg.ConnStr != "" {
		repository.InitTables(ctx, db)

		healthService = service.NewHealthService(repository.NewHealthRepository(db))
		strg = repository.NewMetricsRepository(db)

		logger.Log.Info("Use database storage")
	} else if cfg.File != "" {
		writer, err := fileworker.NewFileWriter(cfg.File)
		if err != nil {
			panic(err)
		}

		strg, err = decorator.NewStoreWithAsyncFile(ctx, storage.NewMemStorage(), time.Duration(cfg.Interval)*time.Second, writer)
		if err != nil {
			panic(err)
		}

		if cfg.Restore {
			err := service.LoadMetrics(cfg.File, strg)
			if err != nil {
				panic(err)
			}
		}
		logger.Log.Info("Use file storage")
	} else {
		strg = storage.NewMemStorage()

		logger.Log.Info("Use mem storage")
	}
	createServer(service.NewMetricsService(strg), healthService, cfg.Host)

}

func createServer(metricsService *service.MetricsService, healthService handler.HealthChecker, host string) {
	r := chi.NewRouter()
	r.With(middleware.LoggingMw, middleware.GzipMW).Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllMetricsHandler(metricsService))
		r.With(middleware.ValidationURLRqMw).Post("/update/{type}/{name}/{value}", handler.MetricsHandler(metricsService))
		r.Post("/update/", handler.JSONUpdateMetricHandler(metricsService))
		r.Post("/updates/", handler.JSONUpdateMetricsHandler(metricsService))
		r.Get("/value/{type}/{name}", handler.GetMetricHandler(metricsService))
		r.Post("/value/", handler.GetJSONMetricHandler(metricsService))
		if healthService != nil {
			r.Get("/ping", handler.HealthHandler(context.TODO(), healthService))
		}
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
