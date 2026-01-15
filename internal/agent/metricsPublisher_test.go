package agent

import (
	"context"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockTempStorage struct {
	mock.Mock
}

func (m *MockTempStorage) GetAndClear() map[string]*models.Metrics {
	args := m.Called()
	return args.Get(0).(map[string]*models.Metrics)
}

// Вспомогательная функция для создания тестовой метрики
func newGauge(name string, value float64) *models.Metrics {
	return &models.Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

func TestMetricsPublisher_Publish_SendsMetricsOnTick(t *testing.T) {
	mockStorage := new(MockTempStorage)
	interval := 100 * time.Millisecond

	publisher, outChan := NewMetricsPublisher(mockStorage, interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := map[string]*models.Metrics{
		"metric1": newGauge("metric1", 123.45),
		"metric2": newGauge("metric2", 678.90),
	}

	mockStorage.On("GetAndClear").Return(metrics).Once()
	mockStorage.On("GetAndClear").Return(map[string]*models.Metrics{}).Maybe()

	go publisher.Publish(ctx)

	received := <-outChan
	assert.Equal(t, metrics[received[0].ID], received[0])
	assert.Equal(t, metrics[received[1].ID], received[1])

	mockStorage.AssertExpectations(t)
}
