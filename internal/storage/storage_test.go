package storage

import (
	"context"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

func Test_memStorage_GetAndClear(t *testing.T) {
	type fields struct {
		storage map[string]*models.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*models.Metrics
	}{
		{
			name: "Clear",
			fields: fields{storage: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			want: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			},
		}, {
			name:   "empty map",
			fields: fields{storage: map[string]*models.Metrics{}},
			want:   map[string]*models.Metrics{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				mutex:   sync.Mutex{},
				storage: tt.fields.storage,
			}
			assert.Equal(t, tt.want, s.GetAndClear())
			assert.Empty(t, s.storage)
		})
	}
}

func Test_memStorage_UpdateMetric(t *testing.T) {
	type fields struct {
		storage map[string]*models.Metrics
	}
	type args struct {
		key   string
		value models.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    map[string]*models.Metrics
	}{
		{
			name: "Add counter1",
			fields: fields{storage: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			args: args{
				key: "counter1",
				value: models.Metrics{
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(58)),
				},
			},
			wantErr: false,
			want: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(100)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			},
		},
		{
			name: "Add gauge1",
			fields: fields{storage: map[string]*models.Metrics{
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			args: args{
				key: "gauge1",
				value: models.Metrics{
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(58.5),
				},
			},
			wantErr: false,
			want: map[string]*models.Metrics{
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(58.5),
				},
			},
		},
		{
			name: "Add new counter2",
			fields: fields{storage: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			args: args{
				key: "counter2",
				value: models.Metrics{
					ID:    "counter2",
					MType: models.Counter,
					Delta: toPtr(int64(25)),
				},
			},
			wantErr: false,
			want: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"counter2": {
					ID:    "counter2",
					MType: models.Counter,
					Delta: toPtr(int64(25)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			},
		},
		{
			name: "wrong type",
			fields: fields{storage: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			args: args{
				key: "counter1",
				value: models.Metrics{
					ID:    "counter1",
					MType: models.Gauge,
					Delta: toPtr(int64(25)),
				},
			},
			wantErr: true,
			want: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				mutex:   sync.Mutex{},
				storage: tt.fields.storage,
			}
			err := s.UpdateMetric(context.TODO(), tt.args.value)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, s.storage)

		})
	}
}

func Test_memStorage_GetAllMetrics(t *testing.T) {
	type fields struct {
		storage map[string]*models.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*models.Metrics
	}{
		{
			name: "not empty",
			fields: fields{storage: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			}},
			want: map[string]*models.Metrics{
				"counter1": {
					ID:    "counter1",
					MType: models.Counter,
					Delta: toPtr(int64(42)),
				},
				"gauge1": {
					ID:    "gauge1",
					MType: models.Gauge,
					Value: toPtr(3.14),
				},
			},
		}, {
			name:   "empty map",
			fields: fields{storage: map[string]*models.Metrics{}},
			want:   map[string]*models.Metrics{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				mutex:   sync.Mutex{},
				storage: tt.fields.storage,
			}
			metrics, err := s.GetAllMetrics(context.TODO())
			assert.Nil(t, err)
			assert.Equal(t, tt.want, metrics)
		})
	}
}

func TestMemStorage_UpdateMetrics(t *testing.T) {
	float64Ptr := func(v float64) *float64 { return &v }
	int64Ptr := func(v int64) *int64 { return &v }

	tests := []struct {
		name         string
		inputMetrics []models.Metrics
		assertions   func(t *testing.T, s *MemStorage)
	}{
		{
			name: "add new gauge metric",
			inputMetrics: []models.Metrics{
				{ID: "NewGauge", MType: models.Gauge, Value: float64Ptr(42.5)},
			},
			assertions: func(t *testing.T, s *MemStorage) {
				metric := s.storage["NewGauge"]
				assert.Equal(t, models.Gauge, metric.MType)
				assert.Equal(t, float64Ptr(42.5), metric.Value)
			},
		},
		{
			name: "add new counter metric",
			inputMetrics: []models.Metrics{
				{ID: "NewCounter", MType: models.Counter, Delta: int64Ptr(100)},
			},
			assertions: func(t *testing.T, s *MemStorage) {
				metric := s.storage["NewCounter"]
				assert.Equal(t, models.Counter, metric.MType)
				assert.Equal(t, int64Ptr(100), metric.Delta)
			},
		},
		{
			name: "update existing gauge",
			inputMetrics: []models.Metrics{
				{ID: "ExistingGauge", MType: models.Gauge, Value: float64Ptr(20.0)},
			},
			assertions: func(t *testing.T, s *MemStorage) {
				metric := s.storage["ExistingGauge"]
				assert.Equal(t, float64Ptr(20.0), metric.Value)
			},
		},
		{
			name: "accumulate counter",
			inputMetrics: []models.Metrics{
				{ID: "AccCounter", MType: models.Counter, Delta: int64Ptr(3)},
			},
			assertions: func(t *testing.T, s *MemStorage) {
				metric := s.storage["AccCounter"]
				assert.Equal(t, int64Ptr(3), metric.Delta)
			},
		},
		{
			name: "multiple metrics mixed",
			inputMetrics: []models.Metrics{
				{ID: "G1", MType: models.Gauge, Value: float64Ptr(200)}, // обновление gauge
				{ID: "C1", MType: models.Counter, Delta: int64Ptr(5)},   // накопление counter
				{ID: "G2", MType: models.Gauge, Value: float64Ptr(300)}, // новая gauge
				{ID: "C2", MType: models.Counter, Delta: int64Ptr(1)},   // новая counter
			},
			assertions: func(t *testing.T, s *MemStorage) {

				metric := s.storage["G1"]
				assert.Equal(t, float64Ptr(200), metric.Value)

				metric = s.storage["C1"]
				assert.Equal(t, int64Ptr(5), metric.Delta)

				metric = s.storage["G2"]
				assert.Equal(t, float64Ptr(300), metric.Value)

				metric = s.storage["C2"]
				assert.Equal(t, int64Ptr(1), metric.Delta)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := NewMemStorage()
			err := s.UpdateMetrics(context.Background(), tt.inputMetrics)
			require.NoError(t, err)
			tt.assertions(t, s)

		})
	}
}

func toPtr[T int64 | float64](value T) *T {
	return &value
}
