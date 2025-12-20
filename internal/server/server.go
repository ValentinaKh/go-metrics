package server

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ValentinaKh/go-metrics/internal/audit/file"
	"github.com/ValentinaKh/go-metrics/internal/audit/rest"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/audit"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/fileworker"
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/handler/middleware"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/repository"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/ValentinaKh/go-metrics/internal/storage/decorator"
)

// ConfigureServer configure server
func ConfigureServer(shutdownCtx context.Context, cfg *config.ServerArg, db *sql.DB) *sync.WaitGroup {
	var strg service.Storage
	var healthService handler.HealthChecker

	if cfg.ConnStr != "" {
		repository.InitTables(shutdownCtx, db)

		healthService = service.NewHealthService(repository.NewHealthRepository(db))

		retryConfig := config.RetryConfig{
			MaxAttempts: 3,
			Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		}
		err := retryConfig.Validate()
		if err != nil {
			panic(err)
		}

		strg = repository.NewMetricsRepository(db,
			retry.NewRetrier(
				retry.NewClassifierRetryPolicy(apperror.NewPostgresErrorClassifier(), retryConfig.MaxAttempts),
				retry.NewStaticDelayStrategy(retryConfig.Delays),
				&retry.SleepTimeProvider{}))

		logger.Log.Info("Use database storage")
	} else if cfg.File != "" {
		writer, err := fileworker.NewFileWriter(cfg.File)
		if err != nil {
			panic(err)
		}

		strg, err = decorator.NewStoreWithAsyncFile(shutdownCtx, storage.NewMemStorage(), time.Duration(cfg.Interval)*time.Second, writer)
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
	auditor := audit.NewAuditor(shutdownCtx, cfg.AuditQueueSize)
	if cfg.AuditFile != "" {
		writer, err := fileworker.NewFileWriter(cfg.AuditFile)
		if err != nil {
			panic(err)
		}
		auditor.Register(file.NewFileAuditHandler(writer))
	}
	if cfg.AuditURL != "" {
		auditor.Register(rest.NewAuditHandler(cfg.AuditURL))
	}
	return createServer(shutdownCtx, service.NewMetricsService(strg), healthService, cfg.Host, cfg.Key, cfg.ProfilePort, auditor)

}

func createServer(ctx context.Context, metricsService *service.MetricsService, healthService handler.HealthChecker,
	host, key, port string, publisher audit.Publisher) *sync.WaitGroup {
	r := chi.NewRouter()
	r.With(middleware.LoggingMw, middleware.DecryptMW, middleware.ValidateHashMW(key), middleware.GzipMW, middleware.HashResponseMW(key)).Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllMetricsHandler(ctx, metricsService))
		r.With(middleware.ValidationURLRqMw).Post("/update/{type}/{name}/{value}", handler.MetricsHandler(ctx, metricsService))
		r.Post("/update/", handler.JSONUpdateMetricHandler(ctx, metricsService, publisher))
		r.Post("/updates/", handler.JSONUpdateMetricsHandler(ctx, metricsService, publisher))
		r.Get("/value/{type}/{name}", handler.GetMetricHandler(ctx, metricsService))
		r.Post("/value/", handler.GetJSONMetricHandler(ctx, metricsService))
		if healthService != nil {
			r.Get("/ping", handler.HealthHandler(ctx, healthService))
		}
	})
	var wg sync.WaitGroup

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.Log.Info("Shutdown timed out, forcing exit")
			} else {
				logger.Log.Error("Error during shutdown: ", zap.Error(err))
			}
		}
	}()

	srvMetrics := &http.Server{
		Addr:    host,
		Handler: r,
	}
	go func() {
		if err := srvMetrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if err := srvMetrics.Shutdown(context.Background()); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.Log.Info("Shutdown timed out, forcing exit")
			} else {
				logger.Log.Error("Error during shutdown: ", zap.Error(err))
			}
		}
	}()
	return &wg
}
