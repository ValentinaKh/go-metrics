package handler

import (
	"encoding/json"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Service interface {
	UpdateMetric(metric models.Metrics) error
	GetMetric(metric models.Metrics) (*models.Metrics, error)
	GetAllMetrics() map[string]string
}

func MetricsHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := parse(chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		errU := service.UpdateMetric(*metric)
		if errU != nil {
			http.Error(w, errU.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func JSONUpdateMetricsHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		errU := service.UpdateMetric(request)
		if errU != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		value, err := service.GetMetric(models.Metrics{ID: name, MType: chi.URLParam(r, "type")})
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

func GetJSONMetricHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		value, err := service.GetMetric(request)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rs, err := json.Marshal(value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(rs)
	}
}

func GetAllMetricsHandler(service Service) http.HandlerFunc {
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
