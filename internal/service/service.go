package service

import (
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"strconv"
)

type Storage interface {
	// UpdateMetric обновляем метрику в хранилище
	UpdateMetric(key string, value models.Metrics) error
	// GetAndClear овозвращаем то, что находится в хранилище и очищаем хранилище
	GetAndClear() map[string]*models.Metrics
	GetAllMetrics() map[string]*models.Metrics
}

type MetricsService struct {
	strg Storage
}

func NewMetricsService(storage Storage) *MetricsService {
	return &MetricsService{strg: storage}
}

func (s MetricsService) Handle(metricType, name, value string) error {
	var metric models.Metrics
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		metric = models.Metrics{
			MType: models.Counter,
			Delta: &value,
		}
	case models.Gauge:
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		metric = models.Metrics{
			MType: models.Gauge,
			Value: &value,
		}
	}
	return s.strg.UpdateMetric(name, metric)
}

func (s MetricsService) GetMetric(name string) (string, bool) {
	metrics := s.strg.GetAllMetrics()
	metric, ok := metrics[name]
	if !ok {
		return "", false
	}
	switch metric.MType {
	case models.Counter:
		return strconv.FormatInt(*metric.Delta, 10), true
	case models.Gauge:
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), true
	}
	return "", false
}

func (s MetricsService) GetAllMetrics() map[string]string {
	result := make(map[string]string)
	for name, metric := range s.strg.GetAllMetrics() {
		var value string
		switch metric.MType {
		case models.Counter:
			value = strconv.FormatInt(*metric.Delta, 10)
		case models.Gauge:
			value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}
		result[name] = value
	}
	return result
}
