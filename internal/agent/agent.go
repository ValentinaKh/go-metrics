package agent

import (
	"context"
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"time"
)

type Agent interface {
	Push(ctx context.Context)
}

type MetricAgent struct {
	s              service.Storage
	h              Sender
	reportInterval time.Duration
}

func NewMetricAgent(s service.Storage, h Sender, reportInterval time.Duration) *MetricAgent {
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
	for _, metric := range metrics {
		rs, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		if err := s.h.Send(rs); err != nil {
			return err
		}
	}
	return nil
}
