package agent

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/config"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/service/collector"
	"github.com/ValentinaKh/go-metrics/internal/service/provider"
	"github.com/ValentinaKh/go-metrics/internal/service/writer"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"sync"
	"time"
)

func ConfigureAgent(shutdownCtx context.Context, cfg *config.AgentArg, rCfg *config.RetryConfig, mChan chan []models.Metrics) *sync.WaitGroup {
	st := storage.NewMemStorage()
	metricPublisher, msgCh := NewMetricsPublisher(st, time.Duration(cfg.ReportInterval)*time.Second)

	var wg sync.WaitGroup
	for idx := 0; idx < int(cfg.RateLimit); idx++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			NewMetricSender(NewPostSender(cfg.Host,
				retry.NewRetrier(
					retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), rCfg.MaxAttempts),
					retry.NewStaticDelayStrategy(rCfg.Delays),
					&retry.SleepTimeProvider{}), cfg.Key), msgCh).
				Push(shutdownCtx)
		}()
	}

	duration := time.Duration(cfg.PollInterval)
	runtimeCollector := collector.NewMetricCollector(provider.NewRuntimeProvider(), duration*time.Second, mChan)
	systemCollector := collector.NewMetricCollector(provider.NewSystemProvider(), duration*time.Second, mChan)
	w := writer.NewMetricWriter(st, mChan)

	//по заданию надо добавить еще одну горутину с новыми метриками
	go func() {
		wg.Add(1)
		defer wg.Done()
		runtimeCollector.Collect(shutdownCtx)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		systemCollector.Collect(shutdownCtx)
	}()

	go metricPublisher.Publish(shutdownCtx)
	go w.Write(shutdownCtx)

	return &wg
}
