package provider

import (
	"math/rand"
	"runtime"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type metricGetter struct {
	name  string
	value func(*runtime.MemStats) float64
}

type RuntimeProvider struct {
	collectors []metricGetter
}

func NewRuntimeProvider() *RuntimeProvider {
	return &RuntimeProvider{
		collectors: runtimeGetters(),
	}
}

func (p *RuntimeProvider) Collect() ([]models.Metrics, error) {
	var runtimeMetrics runtime.MemStats
	runtime.ReadMemStats(&runtimeMetrics)

	var result []models.Metrics

	var count int64
	for _, collect := range p.collectors {
		value := collect.value(&runtimeMetrics)
		result = append(result, models.Metrics{
			ID:    collect.name,
			MType: models.Gauge,
			Value: &value,
		})
		count++
	}

	result = append(result, models.Metrics{
		ID:    models.PollCount,
		MType: models.Counter,
		Delta: &count,
	})

	return result, nil
}

func runtimeGetters() []metricGetter {
	return []metricGetter{
		{models.Alloc, func(ms *runtime.MemStats) float64 { return float64(ms.Alloc) }},
		{models.BuckHashSys, func(ms *runtime.MemStats) float64 { return float64(ms.BuckHashSys) }},
		{models.Frees, func(ms *runtime.MemStats) float64 { return float64(ms.Frees) }},
		{models.GCCPUFraction, func(ms *runtime.MemStats) float64 { return ms.GCCPUFraction }},
		{models.GCSys, func(ms *runtime.MemStats) float64 { return float64(ms.GCSys) }},
		{models.HeapAlloc, func(ms *runtime.MemStats) float64 { return float64(ms.HeapAlloc) }},
		{models.HeapIdle, func(ms *runtime.MemStats) float64 { return float64(ms.HeapIdle) }},
		{models.HeapInuse, func(ms *runtime.MemStats) float64 { return float64(ms.HeapInuse) }},
		{models.HeapObjects, func(ms *runtime.MemStats) float64 { return float64(ms.HeapObjects) }},
		{models.HeapReleased, func(ms *runtime.MemStats) float64 { return float64(ms.HeapReleased) }},
		{models.HeapSys, func(ms *runtime.MemStats) float64 { return float64(ms.HeapSys) }},
		{models.LastGC, func(ms *runtime.MemStats) float64 { return float64(ms.LastGC) }},
		{models.Lookups, func(ms *runtime.MemStats) float64 { return float64(ms.Lookups) }},
		{models.MCacheInuse, func(ms *runtime.MemStats) float64 { return float64(ms.MCacheInuse) }},
		{models.MCacheSys, func(ms *runtime.MemStats) float64 { return float64(ms.MCacheSys) }},
		{models.MSpanInuse, func(ms *runtime.MemStats) float64 { return float64(ms.MSpanInuse) }},
		{models.MSpanSys, func(ms *runtime.MemStats) float64 { return float64(ms.MSpanSys) }},
		{models.Mallocs, func(ms *runtime.MemStats) float64 { return float64(ms.Mallocs) }},
		{models.NextGC, func(ms *runtime.MemStats) float64 { return float64(ms.NextGC) }},
		{models.NumForcedGC, func(ms *runtime.MemStats) float64 { return float64(ms.NumForcedGC) }},
		{models.NumGC, func(ms *runtime.MemStats) float64 { return float64(ms.NumGC) }},
		{models.OtherSys, func(ms *runtime.MemStats) float64 { return float64(ms.OtherSys) }},
		{models.PauseTotalNs, func(ms *runtime.MemStats) float64 { return float64(ms.PauseTotalNs) }},
		{models.StackInuse, func(ms *runtime.MemStats) float64 { return float64(ms.StackInuse) }},
		{models.StackSys, func(ms *runtime.MemStats) float64 { return float64(ms.StackSys) }},
		{models.Sys, func(ms *runtime.MemStats) float64 { return float64(ms.Sys) }},
		{models.TotalAlloc, func(ms *runtime.MemStats) float64 { return float64(ms.TotalAlloc) }},
		{models.RandomValue, func(ms *runtime.MemStats) float64 { return rand.Float64() }},
	}
}
