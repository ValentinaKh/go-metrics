package main

import (
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/handler/middleware"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	err := logger.InitializeZapLogger("info")
	if err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	host := parseArgs()

	r := chi.NewRouter()

	metricsService := service.NewMetricsService(storage.NewMemStorage())

	r.With(middleware.LoggingMw).Get("/", handler.GetAllMetricsHandler(metricsService))
	r.With(middleware.LoggingMw, middleware.ValidationURLRqMw).Post("/update/{type}/{name}/{value}", handler.MetricsHandler(metricsService))
	r.With(middleware.LoggingMw).Post("/update/", handler.JsonUpdateMetricsHandler(metricsService))
	r.With(middleware.LoggingMw).Get("/value/{type}/{name}", handler.GetMetricHandler(metricsService))
	r.With(middleware.LoggingMw).Post("/value/", handler.GetJsonMetricHandler(metricsService))

	return http.ListenAndServe(host, r)

}
