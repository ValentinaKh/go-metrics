package writer

import (
	"context"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/service"
)

// MetricWriter writes metrics to storage
type MetricWriter struct {
	s     service.Storage
	mChan chan []models.Metrics
}

func NewMetricWriter(s service.Storage, mChan chan []models.Metrics) *MetricWriter {
	return &MetricWriter{
		s:     s,
		mChan: mChan,
	}
}

func (w *MetricWriter) Write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close MetricWriter")
			return
		case metrics, ok := <-w.mChan:
			if !ok {
				return
			}
			if err := w.s.UpdateMetrics(ctx, metrics); err != nil {
				logger.Log.Error("Error while save", zap.Error(err))
			}
		}
	}
}
