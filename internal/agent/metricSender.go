package agent

import (
	"context"
	models "github.com/ValentinaKh/go-metrics/internal/model"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
)

// ServerSender - интерфейс для обработки получения метрик из канала
type ServerSender interface {
	Push(ctx context.Context)
}

// Sender - интерфейс для отправки метрик
type Sender interface {
	Close()
	Send(data []*models.Metrics) error
}

type MetricSender struct {
	h     []Sender
	mChan chan []*models.Metrics
}

func NewMetricSender(h []Sender, mChan chan []*models.Metrics) *MetricSender {
	return &MetricSender{h, mChan}
}

// Push - получает метрики из канала и отправляет их на сервер
func (s *MetricSender) Push(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for _, h := range s.h {
				h.Close()
			}
			logger.Log.Info("close MetricSender")
			return
		case msg, ok := <-s.mChan:
			if !ok {
				return
			}
			for _, h := range s.h {
				if err := h.Send(msg); err != nil {
					logger.Log.Error("Error while send", zap.Error(err))
				}
			}
		}
	}
}
