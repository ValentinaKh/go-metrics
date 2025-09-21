package service

import (
	"context"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"math/rand"
	"runtime"
	"time"
)

const (
	alloc         = "Alloc"
	buckHashSys   = "BuckHashSys"
	frees         = "Frees"
	gCCPUFraction = "GCCPUFraction"
	gCSys         = "GCSys"
	heapAlloc     = "HeapAlloc"
	heapIdle      = "HeapIdle"
	heapInuse     = "HeapInuse"
	heapObjects   = "HeapObjects"
	heapReleased  = "HeapReleased"
	heapSys       = "HeapSys"
	lastGC        = "LastGC"
	lookups       = "Lookups"
	mCacheInuse   = "MCacheInuse"
	mCacheSys     = "MCacheSys"
	mSpanInuse    = "MSpanInuse"
	mSpanSys      = "MSpanSys"
	mallocs       = "Mallocs"
	nextGC        = "NextGC"
	numForcedGC   = "NumForcedGC"
	numGC         = "NumGC"
	otherSys      = "OtherSys"
	pauseTotalNs  = "PauseTotalNs"
	stackInuse    = "StackInuse"
	stackSys      = "StackSys"
	sys           = "Sys"
	totalAlloc    = "TotalAlloc"
	pollCount     = "PollCount"
	randomValue   = "RandomValue"
)

type metricGetter struct {
	name  string
	value func(*runtime.MemStats) float64
}

var collectors = []metricGetter{
	{alloc, func(ms *runtime.MemStats) float64 { return float64(ms.Alloc) }},
	{buckHashSys, func(ms *runtime.MemStats) float64 { return float64(ms.BuckHashSys) }},
	{frees, func(ms *runtime.MemStats) float64 { return float64(ms.Frees) }},
	{gCCPUFraction, func(ms *runtime.MemStats) float64 { return ms.GCCPUFraction }},
	{gCSys, func(ms *runtime.MemStats) float64 { return float64(ms.GCSys) }},
	{heapAlloc, func(ms *runtime.MemStats) float64 { return float64(ms.HeapAlloc) }},
	{heapIdle, func(ms *runtime.MemStats) float64 { return float64(ms.HeapIdle) }},
	{heapInuse, func(ms *runtime.MemStats) float64 { return float64(ms.HeapInuse) }},
	{heapObjects, func(ms *runtime.MemStats) float64 { return float64(ms.HeapObjects) }},
	{heapReleased, func(ms *runtime.MemStats) float64 { return float64(ms.HeapReleased) }},
	{heapSys, func(ms *runtime.MemStats) float64 { return float64(ms.HeapSys) }},
	{lastGC, func(ms *runtime.MemStats) float64 { return float64(ms.LastGC) }},
	{lookups, func(ms *runtime.MemStats) float64 { return float64(ms.Lookups) }},
	{mCacheInuse, func(ms *runtime.MemStats) float64 { return float64(ms.MCacheInuse) }},
	{mCacheSys, func(ms *runtime.MemStats) float64 { return float64(ms.MCacheSys) }},
	{mSpanInuse, func(ms *runtime.MemStats) float64 { return float64(ms.MSpanInuse) }},
	{mSpanSys, func(ms *runtime.MemStats) float64 { return float64(ms.MSpanSys) }},
	{mallocs, func(ms *runtime.MemStats) float64 { return float64(ms.Mallocs) }},
	{nextGC, func(ms *runtime.MemStats) float64 { return float64(ms.NextGC) }},
	{numForcedGC, func(ms *runtime.MemStats) float64 { return float64(ms.NumForcedGC) }},
	{numGC, func(ms *runtime.MemStats) float64 { return float64(ms.NumGC) }},
	{otherSys, func(ms *runtime.MemStats) float64 { return float64(ms.OtherSys) }},
	{pauseTotalNs, func(ms *runtime.MemStats) float64 { return float64(ms.PauseTotalNs) }},
	{stackInuse, func(ms *runtime.MemStats) float64 { return float64(ms.StackInuse) }},
	{stackSys, func(ms *runtime.MemStats) float64 { return float64(ms.StackSys) }},
	{sys, func(ms *runtime.MemStats) float64 { return float64(ms.Sys) }},
	{totalAlloc, func(ms *runtime.MemStats) float64 { return float64(ms.TotalAlloc) }},
	{randomValue, func(ms *runtime.MemStats) float64 { return rand.Float64() }},
}

type Collector interface {
	Collect(ctx context.Context)
}

type metricCollector struct {
	s            Storage
	pollInterval time.Duration
}

func NewMetricCollector(s Storage, pollInterval time.Duration) Collector {
	return &metricCollector{
		s:            s,
		pollInterval: pollInterval,
	}
}

func (c *metricCollector) Collect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("close Collector")
			return
		default:
			err := c.collectMetric()

			if err != nil {
				logger.Log.Error(err.Error())
			}

			time.Sleep(c.pollInterval)
		}
	}
}

func (c *metricCollector) addMetric(name string, value float64) error {
	tv := value
	return c.s.UpdateMetric(context.TODO(), models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &tv,
	})
}

func (c *metricCollector) collectMetric() error {

	var runtimeMetrics runtime.MemStats
	runtime.ReadMemStats(&runtimeMetrics)

	var count int64
	for _, collect := range collectors {
		if err := c.addMetric(collect.name, collect.value(&runtimeMetrics)); err != nil {
			return fmt.Errorf("ошибка при сборе метрики: %w", err)
		}
		count++
	}

	if err := c.s.UpdateMetric(context.TODO(), models.Metrics{
		ID:    pollCount,
		MType: models.Counter,
		Delta: &count,
	}); err != nil {
		return fmt.Errorf("ошибка при сборе метрики: %w", err)
	}

	return nil
}
