package agent

import (
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockSender struct {
	Called string
	Err    error
}

func (m *MockSender) Send(data []byte) error {
	if m.Err != nil {
		return m.Err
	}
	m.Called = string(data)
	return nil
}

type MockStorage struct {
	storage map[string]*models.Metrics
}

func (s *MockStorage) GetAndClear() map[string]*models.Metrics {
	return s.storage
}

func Test_metricAgent_send(t *testing.T) {
	type fields struct {
		s              service.TempStorage
		h              Sender
		reportInterval time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{{
		name: "Positive",
		fields: fields{
			s: &MockStorage{storage: map[string]*models.Metrics{
				"intMetric": {
					MType: models.Counter,
					Delta: toPtr(int64(5)),
				},
				"floatMetric": {
					MType: models.Gauge,
					Value: toPtr(5.2),
				},
			}},
			h: &MockSender{
				Err: nil,
			},
			reportInterval: 0,
		},
		want: []string{`{"id":"","type":"gauge","value":5.2}`, `{"id":"","type":"counter","delta":5}`},
	}, {
		name: "Negative",
		fields: fields{
			s: &MockStorage{storage: map[string]*models.Metrics{
				"intMetric": {
					MType: models.Counter,
					Delta: toPtr(int64(5)),
				},
				"floatMetric": {
					MType: models.Gauge,
					Value: toPtr(5.2),
				},
			}},
			h: &MockSender{
				Err: fmt.Errorf("test error"),
			},
			reportInterval: 0,
		},
		want:    []string{},
		wantErr: true,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricAgent{
				s:              tt.fields.s,
				h:              tt.fields.h,
				reportInterval: tt.fields.reportInterval,
			}
			err := s.send()

			mock, ok := tt.fields.h.(*MockSender)
			if !ok {
				t.Fatal("sender is not *MockSender")
			}

			assert.Equal(t, tt.wantErr, err != nil)
			for _, e := range tt.want {
				assert.Contains(t, mock.Called, e)
			}

		})
	}
}
func toPtr[T int64 | float64](value T) *T {
	return &value
}
