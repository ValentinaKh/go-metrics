package service

import (
	"context"
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
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

var collectors = []func(c *metricCollector, runtimeMetrics *runtime.MemStats) error{
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(alloc, float64(runtimeMetrics.Alloc))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(buckHashSys, float64(runtimeMetrics.BuckHashSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(frees, float64(runtimeMetrics.Frees))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(gCCPUFraction, runtimeMetrics.GCCPUFraction)
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(gCSys, float64(runtimeMetrics.GCSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapAlloc, float64(runtimeMetrics.HeapAlloc))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapIdle, float64(runtimeMetrics.HeapIdle))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapInuse, float64(runtimeMetrics.HeapInuse))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapObjects, float64(runtimeMetrics.HeapObjects))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapReleased, float64(runtimeMetrics.HeapReleased))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(heapSys, float64(runtimeMetrics.HeapSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(lastGC, float64(runtimeMetrics.LastGC))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(lookups, float64(runtimeMetrics.Lookups))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(mCacheInuse, float64(runtimeMetrics.MCacheInuse))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(mCacheSys, float64(runtimeMetrics.MCacheSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(mSpanInuse, float64(runtimeMetrics.MSpanInuse))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(mSpanSys, float64(runtimeMetrics.MSpanSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(mallocs, float64(runtimeMetrics.Mallocs))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(nextGC, float64(runtimeMetrics.NextGC))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(numForcedGC, float64(runtimeMetrics.NumForcedGC))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(numGC, float64(runtimeMetrics.NumGC))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(otherSys, float64(runtimeMetrics.OtherSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(pauseTotalNs, float64(runtimeMetrics.PauseTotalNs))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(stackInuse, float64(runtimeMetrics.StackInuse))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(stackSys, float64(runtimeMetrics.StackSys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(sys, float64(runtimeMetrics.Sys))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(totalAlloc, float64(runtimeMetrics.TotalAlloc))
	},
	func(c *metricCollector, runtimeMetrics *runtime.MemStats) error {
		return c.addMetric(randomValue, rand.Float64())
	},
}

type Collector interface {
	Collect(ctx context.Context)
}

type metricCollector struct {
	s            storage.Storage
	pollInterval time.Duration
}

func NewMetricCollector(s storage.Storage, pollInterval time.Duration) Collector {
	return &metricCollector{
		s:            s,
		pollInterval: pollInterval,
	}
}

func (c *metricCollector) Collect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("close Collector")
			return
		default:
			err := c.collectMetric()

			if err != nil {
				fmt.Println(err.Error())
			}

			time.Sleep(c.pollInterval)
		}
	}
}

func (c *metricCollector) addMetric(name string, value float64) error {
	tv := value
	return c.s.UpdateMetric(name, models.Metrics{
		MType: models.Gauge,
		Value: &tv,
	})
}

func (c *metricCollector) collectMetric() error {

	var runtimeMetrics runtime.MemStats
	runtime.ReadMemStats(&runtimeMetrics)

	var count int64
	for _, collect := range collectors {
		if err := collect(c, &runtimeMetrics); err != nil {
			return fmt.Errorf("ошибка при сборе метрики: %w", err)
		}
		count++
	}

	if err := c.s.UpdateMetric(pollCount, models.Metrics{
		MType: models.Counter,
		Delta: &count,
	}); err != nil {
		return fmt.Errorf("ошибка при сборе метрики: %w", err)
	}

	return nil
}
