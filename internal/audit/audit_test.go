package audit

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func float64Ptr(v float64) *float64 { return &v }

type mockObserver struct {
	sync.RWMutex
	metrics []models.Metrics
	ip      string
}

func (m *mockObserver) Update(metrics []models.Metrics, ip string) {
	m.Lock()
	defer m.Unlock()
	m.metrics = metrics
	m.ip = ip
}

func (m *mockObserver) GetMetrics() []models.Metrics {
	m.RLock()
	defer m.RUnlock()
	return m.metrics
}

func (m *mockObserver) GetIP() string {
	m.RLock()
	defer m.RUnlock()
	return m.ip
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

	assert.Equal(t, metrics, observer.GetMetrics())
	assert.Equal(t, ip, observer.GetIP())
}
