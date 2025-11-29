package audit

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// float64Ptr и int64Ptr для удобства
func float64Ptr(v float64) *float64 { return &v }
func int64Ptr(v int64) *int64       { return &v }

// Мок-наблюдатель, который считает вызовы и сохраняет данные
type mockObserver struct {
	metrics []models.Metrics
	ip      string
}

func (m *mockObserver) Update(metrics []models.Metrics, ip string) {
	m.metrics = metrics
	m.ip = ip
}

func TestAuditor_RegisterAndAsyncNotify(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	auditor := NewAuditor(ctx, 10)
	observer := &mockObserver{}
	auditor.Register(observer)

	metrics := []models.Metrics{
		{ID: "TestMetric", MType: "gauge", Value: float64Ptr(42.5)},
	}
	ip := "localhost"

	auditor.Notify(metrics, ip)

	require.Eventually(t, func() bool {
		return len(observer.metrics) == 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	assert.Equal(t, metrics, observer.metrics)
	assert.Equal(t, ip, observer.ip)
}
