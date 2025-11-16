package agent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type TempStorage interface {
	// GetAndClear овозвращаем то, что находится в хранилище и очищаем хранилище
	GetAndClear() map[string]*models.Metrics
}

type Publisher interface {
	Publish(ctx context.Context)
}

type MetricsPublisher struct {
	s              TempStorage
	reportInterval time.Duration
	mChan          chan []byte
}

func NewMetricsPublisher(s TempStorage, reportInterval time.Duration) (*MetricsPublisher, chan []byte) {
	mChan := make(chan []byte, 10)
	return &MetricsPublisher{s: s, reportInterval: reportInterval, mChan: mChan}, mChan
}

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
	rs, err := json.Marshal(request)
	if err != nil {
		return err
	}
	s.mChan <- rs
	return nil
}
