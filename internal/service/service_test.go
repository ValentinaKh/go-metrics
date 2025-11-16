package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type SMockStorage struct {
	storage map[string]*models.Metrics
	err     error
}

func (ms *SMockStorage) UpdateMetric(ctx context.Context, m models.Metrics) error {
	ms.storage[m.ID] = &m
	return ms.err
}
func (ms *SMockStorage) GetAndClear() map[string]*models.Metrics {
	return ms.storage
}
func (ms *SMockStorage) GetAllMetrics(ctx context.Context) (map[string]*models.Metrics, error) {
	return ms.storage, nil
}

func (ms *SMockStorage) UpdateMetrics(ctx context.Context, m []models.Metrics) error {
	for _, metric := range m {
		ms.storage[metric.ID] = &metric
	}
	return ms.err
}

func TestMetricsService_UpdateMetric(t *testing.T) {
	type fields struct {
		s Storage
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
				ID:    "requests",
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
				ID:    "cpu",
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
				ID:    "errors",
				MType: models.Counter,
				Delta: toPtr(int64(-50)),
			}},
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

			service := &MetricsService{
				strg: tt.fields.s,
			}

			metrics, err2 := parse(tt.parts[0], tt.parts[1], tt.parts[2])
			if err2 != nil {
				t.Error(err2)
				return
			}
			err := service.UpdateMetric(context.TODO(), *metrics)

			assert.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr {
				assert.Equal(t, tt.want, service.strg.(*SMockStorage).storage)
			}
		})
	}
}

func TestMetricsService_GetMetric(t *testing.T) {
	type fields struct {
		s Storage
	}
	tests := []struct {
		name       string
		fields     fields
		metricName string
		metricType string
		exist      bool
		want       *models.Metrics
	}{
		{
			name: "counter metric exists",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"requests": {
					ID:    "requests",
					MType: models.Counter,
					Delta: toPtr(int64(100)),
				}},
				err: nil},
			},
			metricName: "requests",
			metricType: models.Counter,
			exist:      true,
			want: &models.Metrics{
				ID:    "requests",
				MType: models.Counter,
				Delta: toPtr(int64(100)),
			},
		},
		{
			name: "gauge metric exists",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"cpu": {
					ID:    "cpu",
					MType: models.Gauge,
					Value: toPtr(0.85),
				}},
				err: nil},
			},
			metricName: "cpu",
			metricType: models.Gauge,
			exist:      true,
			want: &models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
				Value: toPtr(0.85),
			},
		},
		{
			name: "unknown type",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{"cpu": {
					ID:    "cpu",
					MType: "counter",
					Value: toPtr(0.85),
				}},
				err: nil},
			},
			metricName: "cpu",
			metricType: "unknown",
			exist:      false,
			want:       nil,
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
			metricType: models.Gauge,
			exist:      false,
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			service := &MetricsService{
				strg: tt.fields.s,
			}

			metric, err := service.GetMetric(context.TODO(), models.Metrics{
				ID:    tt.metricName,
				MType: tt.metricType,
			})

			assert.Equal(t, tt.exist, err == nil)
			assert.Equal(t, tt.want, metric)
		})
	}
}

func TestMetricsService_GetAllMetrics(t *testing.T) {
	type fields struct {
		s Storage
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

			service := &MetricsService{
				strg: tt.fields.s,
			}

			metrics, err := service.GetAllMetrics(context.TODO())
			assert.Nil(t, err)
			assert.Equal(t, tt.want, metrics)
		})
	}
}

func toPtr[T int64 | float64](value T) *T {
	return &value
}

func parse(metricType, name, value string) (*models.Metrics, error) {
	var metric models.Metrics
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		}
	case models.Gauge:
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}
	}
	return &metric, nil
}
