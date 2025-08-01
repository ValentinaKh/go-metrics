package main

import (
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/handler/middleware"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()

	mux.Handle("/", middleware.ValidationPostMw(middleware.ValidationUrlRqMw(handler.MetricsHandler(service.NewMetricsService()))))
	return http.ListenAndServe(`:8080`, mux)
}
