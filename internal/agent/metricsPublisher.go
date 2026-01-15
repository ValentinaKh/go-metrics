package agent

import (
	"context"
	"time"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

// TempStorage - интерфейс для хранилища метрик
type TempStorage interface {
	// GetAndClear овозвращаем то, что находится в хранилище и очищаем хранилище
	GetAndClear() map[string]*models.Metrics
}

type Publisher interface {
	Publish(ctx context.Context)
}

// MetricsPublisher - содержит используемое хранилище, а так же канал для отправки метрик
type MetricsPublisher struct {
	s              TempStorage
	reportInterval time.Duration
	mChan          chan []*models.Metrics
}

func NewMetricsPublisher(s TempStorage, reportInterval time.Duration) (*MetricsPublisher, chan []*models.Metrics) {
	mChan := make(chan []*models.Metrics, 10)
	return &MetricsPublisher{s: s, reportInterval: reportInterval, mChan: mChan}, mChan
}

// Publish - забирает из хранилища метрики и отправляет их в канал для последующей отправки на сервер
func (s *MetricsPublisher) Publish(ctx context.Context) {
	ticker := time.NewTicker(s.reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close MetricsPublisher")
			close(s.mChan)
			return
		case <-ticker.C:
			err := s.send()
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}
}

func (s *MetricsPublisher) send() error {
	metrics := s.s.GetAndClear()
	if len(metrics) == 0 {
		return nil
	}
	request := make([]*models.Metrics, 0)
	for _, m := range metrics {
		request = append(request, m)
	}
	s.mChan <- request
	return nil
}
