package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/audit"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type Service interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
	GetMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error)
	GetAllMetrics(ctx context.Context) (map[string]string, error)
}

func MetricsHandler(ctx context.Context, service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		metric, err := parse(chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value"))
		if err != nil {
			logger.Log.Error("UpdateMetric", zap.Error(err))

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		errU := service.UpdateMetric(timeout, *metric)
		if errU != nil {
			logger.Log.Error("UpdateMetric", zap.Error(errU))

			http.Error(w, errU.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func JSONUpdateMetricHandler(ctx context.Context, service Service, p audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var request models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		errU := service.UpdateMetric(timeout, request)
		if errU != nil {
			logger.Log.Error("UpdateMetric", zap.Error(errU))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		p.Notify([]models.Metrics{request}, r.URL.Host)
	}
}

func GetMetricHandler(ctx context.Context, service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		name := chi.URLParam(r, "name")
		value, err := service.GetMetric(timeout, models.Metrics{ID: name, MType: chi.URLParam(r, "type")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		var v string
		switch value.MType {
		case models.Counter:
			v = strconv.FormatInt(*value.Delta, 10)
		case models.Gauge:
			v = strconv.FormatFloat(*value.Value, 'f', -1, 64)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set(name, v)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(v))
	}
}

func GetJSONMetricHandler(ctx context.Context, service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		var request models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		value, err := service.GetMetric(timeout, request)
		if err != nil {
			logger.Log.Error("GetMetric", zap.Error(err))

			w.WriteHeader(http.StatusNotFound)
			return
		}

		rs, err := json.Marshal(value)
		if err != nil {
			logger.Log.Error("GetMetric", zap.Error(err))

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(rs)
	}
}

func GetAllMetricsHandler(ctx context.Context, service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		values, err := service.GetAllMetrics(timeout)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

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

func JSONUpdateMetricsHandler(ctx context.Context, service Service, p audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var request []models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Error("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		errU := service.UpdateMetrics(timeout, request)
		if errU != nil {
			logger.Log.Error("UpdateMetrics", zap.Error(errU))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		p.Notify(request, r.URL.Host)
	}
}

func parse(metricType, name, value string) (*models.Metrics, error) {
	var metric models.Metrics
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		}
	case models.Gauge:
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}
	default:
		return nil, fmt.Errorf("неизвестный тип метрики %s", metricType)
	}
	return &metric, nil
}
