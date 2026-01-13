package audit

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func float64Ptr(v float64) *float64 { return &v }

type mockObserver struct {
	updates chan task
}

func newMockObserver() *mockObserver {
	return &mockObserver{
		updates: make(chan task, 1),
	}
}
func (m *mockObserver) Update(metrics []models.Metrics, ip string) {
	m.updates <- task{metrics, ip}
}

func (m *mockObserver) AwaitUpdate(timeout time.Duration) (*task, bool) {
	select {
	case res := <-m.updates:
		return &res, true
	case <-time.After(timeout):
		return nil, false
	}
}

func TestAuditor_RegisterAndAsyncNotify(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	auditor := NewAuditor(ctx, 10)
	observer := newMockObserver()
	auditor.Register(observer)

	metrics := []models.Metrics{
		{ID: "TestMetric", MType: "gauge", Value: float64Ptr(42.5)},
	}
	ip := "localhost"

	auditor.Notify(metrics, ip)

	task, ok := observer.AwaitUpdate(500 * time.Millisecond)
	require.True(t, ok)
	assert.Equal(t, metrics, task.metrics)
	assert.Equal(t, ip, task.ip)
}
