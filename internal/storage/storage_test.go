package storage

import (
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
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
			s := &memStorage{
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
			s := &memStorage{
				mutex:   sync.Mutex{},
				storage: tt.fields.storage,
			}
			err := s.UpdateMetric(tt.args.key, tt.args.value)
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
			s := &memStorage{
				mutex:   sync.Mutex{},
				storage: tt.fields.storage,
			}
			assert.Equal(t, tt.want, s.GetAllMetrics())
		})
	}
}

func toPtr[T int64 | float64](value T) *T {
	return &value
}
