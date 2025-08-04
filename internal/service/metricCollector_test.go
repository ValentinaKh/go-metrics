package service

import (
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

type MockStorage struct {
	storage          map[string]*models.Metrics
	UpdateMetricFunc func(name string, m models.Metrics) error
}

func (ms *MockStorage) UpdateMetric(name string, m models.Metrics) error {
	ms.storage[name] = &m
	return ms.UpdateMetricFunc(name, m)
}

func (ms *MockStorage) GetAndClear() map[string]*models.Metrics {
	return ms.storage
}

func (ms *MockStorage) GetAllMetrics() map[string]*models.Metrics {
	return nil
}

func Test_metricCollector_addMetric(t *testing.T) {
	type fields struct {
		s            Storage
		pollInterval time.Duration
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive test",
			fields: fields{
				s: &MockStorage{
					storage: map[string]*models.Metrics{},
					UpdateMetricFunc: func(name string, m models.Metrics) error {
						return nil
					}},
				pollInterval: 2,
			},
			args: args{
				name:  "metric",
				value: 5.1,
			},
			wantErr: false,
		},
		{
			name: "negative test",
			fields: fields{
				s: &MockStorage{
					storage: map[string]*models.Metrics{},
					UpdateMetricFunc: func(name string, m models.Metrics) error {
						return fmt.Errorf("test error")
					}},
				pollInterval: 2,
			},
			args: args{
				name:  "metric",
				value: 5.1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &metricCollector{
				s:            tt.fields.s,
				pollInterval: tt.fields.pollInterval,
			}

			assert.Equal(t, tt.wantErr, c.addMetric(tt.args.name, tt.args.value) != nil)
		})
	}
}

func Test_metricCollector_collectMetric(t *testing.T) {
	originalCollectors := collectors
	defer func() { collectors = originalCollectors }()

	type fields struct {
		s             Storage
		pollInterval  time.Duration
		tmpCollectors []func(*metricCollector, *runtime.MemStats) error
	}
	tests := []struct {
		name    string
		fields  fields
		calls   int64
		wantErr bool
	}{
		{
			name: "positive test",
			fields: fields{
				s: &MockStorage{
					storage: map[string]*models.Metrics{},
					UpdateMetricFunc: func(name string, m models.Metrics) error {
						return nil
					}},
				pollInterval: 2,
				tmpCollectors: []func(*metricCollector, *runtime.MemStats) error{
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
				},
			},
			calls:   3,
			wantErr: false,
		},
		{
			name: "add return error",
			fields: fields{
				s: &MockStorage{
					storage: map[string]*models.Metrics{},
					UpdateMetricFunc: func(name string, m models.Metrics) error {
						return nil
					}},
				pollInterval: 2,
				tmpCollectors: []func(*metricCollector, *runtime.MemStats) error{
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
					func(c *metricCollector, m *runtime.MemStats) error { return fmt.Errorf("test error") },
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
				},
			},
			calls:   0,
			wantErr: true,
		},
		{
			name: "update counter metric return error",
			fields: fields{
				s: &MockStorage{
					storage: map[string]*models.Metrics{},
					UpdateMetricFunc: func(name string, m models.Metrics) error {
						return fmt.Errorf("add error")
					}},
				pollInterval: 2,
				tmpCollectors: []func(*metricCollector, *runtime.MemStats) error{
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
					func(c *metricCollector, m *runtime.MemStats) error { return nil },
				},
			},
			calls:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		c := &metricCollector{
			s:            tt.fields.s,
			pollInterval: tt.fields.pollInterval,
		}
		collectors = tt.fields.tmpCollectors

		err := c.collectMetric()
		assert.Equal(t, tt.wantErr, err != nil)
		if err == nil {
			assert.Equal(t, *tt.fields.s.(*MockStorage).storage[pollCount].Delta, tt.calls)
		}

	}
}
