package agent

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type ServerSender interface {
	Push(ctx context.Context)
}

type Sender interface {
	Send(data []byte) error
}

type MetricSender struct {
	h     Sender
	mChan chan []byte
}

func NewMetricSender(h Sender, mChan chan []byte) *MetricSender {
	return &MetricSender{h, mChan}
}

func (s *MetricSender) Push(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close MetricSender")
			return
		case msg, ok := <-s.mChan:
			if !ok {
				return
			}
			if err := s.h.Send(msg); err != nil {
				logger.Log.Error("Error while send", zap.Error(err))
			}
		}
	}
}
