package service

import (
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SMockStorage struct {
	storage map[string]*models.Metrics
	err     error
}

func (ms *SMockStorage) UpdateMetric(name string, m models.Metrics) error {
	ms.storage[name] = &m
	return ms.err
}
func (ms *SMockStorage) GetAndClear() map[string]*models.Metrics {
	return ms.storage
}
func (ms *SMockStorage) GetAllMetrics() map[string]*models.Metrics {
	return ms.storage
}

func TestMetricsService_Handle(t *testing.T) {
	type fields struct {
		s storage.Storage
	}
	tests := []struct {
		name    string
		parts   []string
		fields  fields
		wantErr bool
		want    map[string]*models.Metrics
	}{
		{
			name:  "valid counter",
			parts: []string{"counter", "requests", "100"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: false,
			want: map[string]*models.Metrics{"requests": {
				MType: models.Counter,
				Delta: toPtr(int64(100)),
			}},
		},
		{
			name:  "valid gauge",
			parts: []string{"gauge", "cpu", "0.85"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: false,
			want: map[string]*models.Metrics{"cpu": {
				MType: models.Gauge,
				Value: toPtr(0.85),
			}},
		},
		{
			name:  "counter with negative value",
			parts: []string{"counter", "errors", "-50"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: false,
			want: map[string]*models.Metrics{"errors": {
				MType: models.Counter,
				Delta: toPtr(int64(-50)),
			}},
		},
		{
			name:  "invalid counter value",
			parts: []string{"counter", "requests", "abc"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: true,
			want:    map[string]*models.Metrics{},
		},
		{
			name:  "invalid gauge value",
			parts: []string{"gauge", "cpu", "invalid"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: true,
			want:    map[string]*models.Metrics{},
		},
		{
			name:  "storage returns error",
			parts: []string{"counter", "requests", "100"},
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     fmt.Errorf("test error")},
			},
			wantErr: true,
			want:    map[string]*models.Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			service := &metricsService{
				strg: tt.fields.s,
			}

			err := service.Handle(tt.parts[0], tt.parts[1], tt.parts[2])

			assert.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr {
				assert.Equal(t, tt.want, service.strg.(*SMockStorage).storage)
			}
		})
	}
}

func TestMetricsService_GetMetric(t *testing.T) {
	type fields struct {
		s storage.Storage
	}
	tests := []struct {
		name       string
		fields     fields
		metricName string
		exist      bool
		want       string
	}{
		{
			name: "counter metric exists",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"requests": {
					MType: models.Counter,
					Delta: toPtr(int64(100)),
				}},
				err: nil},
			},
			metricName: "requests",
			exist:      true,
			want:       "100",
		},
		{
			name: "gauge metric exists",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"cpu": {
					MType: models.Gauge,
					Value: toPtr(0.85),
				}},
				err: nil},
			},
			metricName: "cpu",
			exist:      true,
			want:       "0.85",
		},
		{
			name: "unknown type",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"cpu": {
					MType: "unknown",
					Value: toPtr(0.85),
				}},
				err: nil},
			},
			metricName: "cpu",
			exist:      false,
			want:       "",
		},
		{
			name: "unknown metric",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"cpu": {
					MType: models.Gauge,
					Value: toPtr(0.85),
				}},
				err: nil},
			},
			metricName: "memory",
			exist:      false,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			service := &metricsService{
				strg: tt.fields.s,
			}

			metric, ok := service.GetMetric(tt.metricName)

			assert.Equal(t, tt.exist, ok)
			assert.Equal(t, tt.want, metric)
		})
	}
}

func TestMetricsService_GetAllMetrics(t *testing.T) {
	type fields struct {
		s storage.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "get all values",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"requests": {
					MType: models.Counter,
					Delta: toPtr(int64(100)),
				},
					"cpu": {
						MType: models.Gauge,
						Value: toPtr(0.85),
					}},
				err: nil},
			},

			want: map[string]string{"requests": "100", "cpu": "0.85"},
		},
		{
			name: "empty map",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},

			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			service := &metricsService{
				strg: tt.fields.s,
			}

			assert.Equal(t, tt.want, service.GetAllMetrics())
		})
	}
}

func toPtr[T int64 | float64](value T) *T {
	return &value
}
