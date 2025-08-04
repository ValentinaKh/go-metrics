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
	CalledWithURL []string
	Err           error
}

func (m *MockSender) Send(url string) error {
	if m.Err != nil {
		return m.Err
	}
	m.CalledWithURL = append(m.CalledWithURL, url)
	return nil
}

type MockStorage struct {
	storage map[string]*models.Metrics
}

func (s *MockStorage) UpdateMetric(_ string, _ models.Metrics) error {
	return nil
}

func (s *MockStorage) GetAndClear() map[string]*models.Metrics {
	return s.storage
}

func (s *MockStorage) GetAllMetrics() map[string]*models.Metrics {
	return s.storage
}

func Test_metricAgent_send(t *testing.T) {
	type fields struct {
		s              service.Storage
		h              Sender
		reportInterval time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantURL []string
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
				CalledWithURL: make([]string, 0),
				Err:           nil,
			},
			reportInterval: 0,
		},
		wantURL: []string{"http://localhost:8080/update/counter/intMetric/5", "http://localhost:8080/update/gauge/floatMetric/5.2"},
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
				CalledWithURL: make([]string, 0),
				Err:           fmt.Errorf("test error"),
			},
			reportInterval: 0,
		},
		wantURL: []string{},
		wantErr: true,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricAgent{
				s:              tt.fields.s,
				h:              tt.fields.h,
				reportInterval: tt.fields.reportInterval,
				host:           "localhost:8080",
			}
			err := s.send()

			mock, ok := tt.fields.h.(*MockSender)
			if !ok {
				t.Fatal("sender is not *MockSender")
			}

			assert.Equal(t, tt.wantErr, err != nil)
			assert.ElementsMatch(t, tt.wantURL, mock.CalledWithURL)
		})
	}
}
func toPtr[T int64 | float64](value T) *T {
	return &value
}
