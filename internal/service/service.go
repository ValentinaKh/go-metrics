package service

import (
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"strconv"
)

type MetricsService struct {
	strg storage.Storage
}

func NewMetricsService() *MetricsService {
	return &MetricsService{strg: storage.NewMemStorage()}
}

func (s MetricsService) Handle(url string) error {
	parts := utils.ParseUrl(url)
	var metric models.Metrics
	switch parts[1] {
	case models.Counter:
		value, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return err
		}
		metric = models.Metrics{
			MType: models.Counter,
			Delta: &value,
		}
	case models.Gauge:
		value, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return err
		}
		metric = models.Metrics{
			MType: models.Gauge,
			Value: &value,
		}
	}
	return s.strg.UpdateMetric(parts[2], metric)
}
