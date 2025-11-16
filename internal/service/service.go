package service

import (
	"context"
	"fmt"
	"strconv"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type Storage interface {
	// UpdateMetric обновляем метрику в хранилище
	UpdateMetric(ctx context.Context, value models.Metrics) error
	UpdateMetrics(ctx context.Context, values []models.Metrics) error
	GetAllMetrics(ctx context.Context) (map[string]*models.Metrics, error)
}

type MetricsService struct {
	strg Storage
}

func NewMetricsService(storage Storage) *MetricsService {
	return &MetricsService{strg: storage}
}

func (s MetricsService) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	return s.strg.UpdateMetric(ctx, metric)
}

func (s MetricsService) GetMetric(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	metrics, err := s.strg.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения метрик %w", err)
	}
	metric, ok := metrics[m.ID]
	if !ok {
		return nil, fmt.Errorf("метрика не найдена")
	}
	if metric.MType != m.MType {
		return nil, fmt.Errorf("метрика с таким типом не найдена")
	}
	return metric, nil
}

func (s MetricsService) GetAllMetrics(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	metrics, err := s.strg.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения метрик %w", err)
	}

	for name, metric := range metrics {
		var value string
		switch metric.MType {
		case models.Counter:
			value = strconv.FormatInt(*metric.Delta, 10)
		case models.Gauge:
			value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}
		result[name] = value
	}
	return result, nil
}

func (s MetricsService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	return s.strg.UpdateMetrics(ctx, metrics)
}
