package agent

import (
	"context"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
)

// ServerSender - интерфейс для обработки получения метрик из канала
type ServerSender interface {
	Push(ctx context.Context)
}

// Sender - интерфейс для отправки метрик
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

// Push - получает метрики из канала и отправляет их на сервер
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
