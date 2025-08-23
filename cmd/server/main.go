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
	logger.Setup("info")

	host := parseArgs()

	r := chi.NewRouter()

	metricsService := service.NewMetricsService(storage.NewMemStorage())

	r.Get("/", handler.GetAllMetricsHandler(metricsService))
	r.With(middleware.ValidationURLRqMw).Post("/update/{type}/{name}/{value}", handler.MetricsHandler(metricsService))
	r.Get("/value/{type}/{name}", handler.GetMetricHandler(metricsService))

	return http.ListenAndServe(host, r)

}
