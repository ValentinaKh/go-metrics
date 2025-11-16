package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	models "github.com/ValentinaKh/go-metrics/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricProvider struct {
	mock.Mock
}

func (m *MockMetricProvider) Collect() ([]models.Metrics, error) {
	args := m.Called()
	return args.Get(0).([]models.Metrics), args.Error(1)
}

func newGauge(name string, value float64) models.Metrics {
	return models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
}

func TestMetricCollector_Collect_SendsMetricsOnTick(t *testing.T) {
	mockProvider := new(MockMetricProvider)
	interval := 100 * time.Millisecond

	mChan := make(chan []models.Metrics, 10)
	defer close(mChan)
	collector := NewMetricCollector(mockProvider, interval, mChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := []models.Metrics{
		newGauge("cpu", 45.5),
		newGauge("mem", 1024.0),
	}

	mockProvider.On("Collect").Return(metrics, nil).Once()
	mockProvider.On("Collect").Return([]models.Metrics{}, nil).Maybe()

	go collector.Collect(ctx)

	received := <-mChan

	assert.Len(t, received, 2)
	assert.Equal(t, "cpu", received[0].ID)
	assert.Equal(t, 45.5, *received[0].Value)
	assert.Equal(t, "mem", received[1].ID)
	assert.Equal(t, 1024.0, *received[1].Value)

	mockProvider.AssertExpectations(t)
}

func TestMetricCollector_Collect_ContinuesAfterError(t *testing.T) {
	mockProvider := new(MockMetricProvider)
	interval := 100 * time.Millisecond

	mChan := make(chan []models.Metrics, 10)
	defer close(mChan)
	collector := NewMetricCollector(mockProvider, interval, mChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := []models.Metrics{
		newGauge("cpu", 45.5),
		newGauge("mem", 1024.0),
	}

	mockProvider.On("Collect").Return(metrics, nil)
	mockProvider.On("Collect").Return(nil, errors.New("send failed"))
	mockProvider.On("Collect").Return(metrics, nil)
	mockProvider.On("Collect").Return([]models.Metrics{}, nil).Maybe()

	go collector.Collect(ctx)

	received := <-mChan
	received2 := <-mChan

	assert.Len(t, received, 2)
	assert.Equal(t, "cpu", received[0].ID)
	assert.Equal(t, 45.5, *received[0].Value)
	assert.Equal(t, "mem", received[1].ID)
	assert.Equal(t, 1024.0, *received[1].Value)

	assert.Len(t, received2, 2)
	assert.Equal(t, "cpu", received2[0].ID)
	assert.Equal(t, 45.5, *received2[0].Value)
	assert.Equal(t, "mem", received2[1].ID)
	assert.Equal(t, 1024.0, *received2[1].Value)

	mockProvider.AssertExpectations(t)
}
