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

func TestMetricsService_Handle(t *testing.T) {
	type fields struct {
		s storage.Storage
	}
	tests := []struct {
		name    string
		url     string
		fields  fields
		wantErr bool
		want    map[string]*models.Metrics
	}{
		{
			name: "valid counter",
			url:  "/update/counter/requests/100",
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
			name: "valid gauge",
			url:  "/update/gauge/cpu/0.85",
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
			name: "counter with negative value",
			url:  "/update/counter/errors/-50",
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
			name: "invalid counter value",
			url:  "/update/counter/requests/abc",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: true,
			want:    map[string]*models.Metrics{},
		},
		{
			name: "invalid gauge value",
			url:  "/update/gauge/cpu/invalid",
			fields: fields{s: &SMockStorage{
				storage: map[string]*models.Metrics{},
				err:     nil},
			},
			wantErr: true,
			want:    map[string]*models.Metrics{},
		},
		{
			name: "storage returns error",
			url:  "/update/counter/requests/100",
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

			err := service.Handle(tt.url)

			assert.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr {
				assert.Equal(t, tt.want, service.strg.(*SMockStorage).storage)
			}
		})
	}
}

func toPtr[T int64 | float64](value T) *T {
	return &value
}
