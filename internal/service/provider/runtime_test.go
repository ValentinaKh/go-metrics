package provider

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

func TestRuntimeProvider_Collect_WithManualProvider(t *testing.T) {

	provider := &RuntimeProvider{
		collectors: []metricGetter{
			{models.Alloc, func(ms *runtime.MemStats) float64 { return 1024.0 }},
			{models.HeapAlloc, func(ms *runtime.MemStats) float64 { return 512.0 }},
		},
	}

	metrics, err := provider.Collect()
	require.NoError(t, err)
	require.NotNil(t, metrics)

	assert.Equal(t, 3, len(metrics))
	count := int64(2)
	assert.Equal(t, []models.Metrics{newGauge(models.Alloc, 1024.0), newGauge(models.HeapAlloc, 512), models.Metrics{
		ID:    models.PollCount,
		MType: models.Counter,
		Delta: &count,
	}}, metrics)

}

func newGauge(name string, value float64) models.Metrics {
	return models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
}
