package agent

import (
	"context"
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"time"
)

type Agent interface {
	Push(ctx context.Context)
}

type MetricAgent struct {
	s              service.TempStorage
	h              Sender
	reportInterval time.Duration
}

func NewMetricAgent(s service.TempStorage, h Sender, reportInterval time.Duration) *MetricAgent {
	return &MetricAgent{s, h, reportInterval}
}

func (s *MetricAgent) Push(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close agent")
			return
		default:
			err := s.send()
			if err != nil {
				logger.Log.Error(err.Error())
			}
			time.Sleep(s.reportInterval)
		}
	}
}

func (s *MetricAgent) send() error {
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
	if err := s.h.Send(rs); err != nil {
		return err
	}
	return nil
}
