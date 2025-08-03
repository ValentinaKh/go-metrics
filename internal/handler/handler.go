package handler

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func MetricsHandler(service service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := service.Handle(chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricHandler(service service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		value, ok := service.GetMetric(name)
		if !ok {
			http.Error(w, "Метрика не найдена", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set(name, value)
		w.Write([]byte(value))
	}
}

func GetAllMetricsHandler(service service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := service.GetAllMetrics()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Metrics</title></head><body>
<h1>Metrics</h1>
<ul>`))

		for name, m := range values {
			fmt.Fprintf(w, `<li><strong>%s</strong> %s</li>`, name, m)
		}
		w.Write([]byte(`</ul></body></html>`))
	}
}
