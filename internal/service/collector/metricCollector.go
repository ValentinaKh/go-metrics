package collector

import (
	"context"
	"time"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

// Collector is an interface for collecting metrics
type Collector interface {
	Collect(ctx context.Context)
}

type MetricProvider interface {
	Collect() ([]models.Metrics, error)
}

type metricCollector struct {
	pollInterval time.Duration
	provider     MetricProvider
	mChan        chan []models.Metrics
}

func NewMetricCollector(provider MetricProvider, pollInterval time.Duration, mChan chan []models.Metrics) Collector {
	return &metricCollector{
		provider:     provider,
		pollInterval: pollInterval,
		mChan:        mChan,
	}
}

// Collect collects metrics from the provider
func (c *metricCollector) Collect(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close Collector")
			return
		case <-ticker.C:
			metrics, err := c.provider.Collect()
			if err != nil {
				logger.Log.Error(err.Error())
			} else {
				c.mChan <- metrics
			}
		}
	}
}
