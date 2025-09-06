package service

import (
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"strconv"
)

type Storage interface {
	// UpdateMetric обновляем метрику в хранилище
	UpdateMetric(value models.Metrics) error
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

func (s MetricsService) UpdateMetric(metric models.Metrics) error {
	return s.strg.UpdateMetric(metric)
}

func (s MetricsService) GetMetric(m models.Metrics) (*models.Metrics, error) {
	metrics := s.strg.GetAllMetrics()
	metric, ok := metrics[m.ID]
	if !ok {
		return nil, fmt.Errorf("метрика не найдена")
	}
	if metric.MType != m.MType {
		return nil, fmt.Errorf("метрика с таким типом не найдена")
	}
	return metric, nil
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
