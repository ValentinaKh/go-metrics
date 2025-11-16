package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ValentinaKh/go-metrics/internal/audit"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type MockMetricsService struct {
	HandleFunc        func(metric models.Metrics) error
	GetMetricFunc     func(metric models.Metrics) (*models.Metrics, error)
	GetAllMetricsFunc func() map[string]string
}

func (m *MockMetricsService) UpdateMetric(_ context.Context, metric models.Metrics) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(metric)
	}
	return nil
}

func (m *MockMetricsService) GetMetric(_ context.Context, metric models.Metrics) (*models.Metrics, error) {
	return m.GetMetricFunc(metric)
}

func (m *MockMetricsService) GetAllMetrics(_ context.Context) (map[string]string, error) {
	return m.GetAllMetricsFunc(), nil
}

func (m *MockMetricsService) UpdateMetrics(_ context.Context, metrics []models.Metrics) error {
	return nil
}

func TestMetricsHandler(t *testing.T) {
	type args struct {
		service Service
		url     string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return nil
				},
			},
				url: "/update/gauge/cpu_load/0.85",
			},
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name: "unknown metric",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return nil
				},
			},
				url: "/update/unknown/cpu_load/0.85",
			},
			want: want{
				code:     400,
				response: "неизвестный тип метрики unknown\n",
			},
		},
		{
			name: "not found",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return fmt.Errorf("not found")
				},
			},
				url: "/update/counter/cpu/1",
			},
			want: want{
				code:     400,
				response: "not found\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := MetricsHandler(context.TODO(), test.args.service)
			request := httptest.NewRequest(http.MethodPost, test.args.url, nil)
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler)

			r.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))

		})
	}
}

func Test_GetMetricHandler(t *testing.T) {
	var res int64 = 500
	type args struct {
		service Service
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test",
			args: args{&MockMetricsService{
				GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
					return &models.Metrics{
						ID:    metric.ID,
						MType: metric.MType,
						Delta: &res,
						Value: metric.Value,
						Hash:  "",
					}, nil
				},
			},
			},
			want: want{
				code:     200,
				response: "500",
			},
		},
		{
			name: "not found",
			args: args{&MockMetricsService{
				GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
					return nil, fmt.Errorf("метрика не найдена")
				},
			},
			},
			want: want{
				code:     404,
				response: "метрика не найдена\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := GetMetricHandler(context.TODO(), test.args.service)
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", handler)
			server := httptest.NewServer(r)
			defer server.Close()

			resp, err := server.Client().Get(server.URL + "/value/counter/custom")
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)

			body := resp.Body
			data, _ := io.ReadAll(body)

			assert.Equal(t, test.want.response, string(data))
		})
	}
}

func Test_GetAllMetricsHandler(t *testing.T) {
	type args struct {
		service Service
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test",
			args: args{&MockMetricsService{
				GetAllMetricsFunc: func() map[string]string {
					return map[string]string{"cpu": "0.54"}
				},
			},
			},
			want: want{
				code:     200,
				response: "<!DOCTYPE html>\n<html><head><title>Metrics</title></head><body>\n<h1>Metrics</h1>\n<ul><li><strong>cpu</strong> 0.54</li></ul></body></html>",
			},
		},
		{
			name: "empty map",
			args: args{&MockMetricsService{
				GetAllMetricsFunc: func() map[string]string {
					return map[string]string{}
				},
			},
			},
			want: want{
				code:     200,
				response: "<!DOCTYPE html>\n<html><head><title>Metrics</title></head><body>\n<h1>Metrics</h1>\n<ul></ul></body></html>",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := GetAllMetricsHandler(context.TODO(), test.args.service)
			r := chi.NewRouter()
			r.Get("/", handler)
			server := httptest.NewServer(r)
			defer server.Close()

			resp, err := server.Client().Get(server.URL)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)

			body := resp.Body
			data, _ := io.ReadAll(body)

			assert.Equal(t, test.want.response, string(data))
		})
	}
}

func TestJsonUpdateMetricHandler(t *testing.T) {
	type args struct {
		service Service
		json    string
	}
	type want struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return nil
				},
			},
				json: `{"id": "LastGC","type": "gauge","value": 1744184459}`,
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "decode error",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return nil
				},
			},
				json: `{
  						"id": "LastGC",
  						"type": "gauge",
  						"value": 1744184459
					`,
			},
			want: want{
				code: 500,
			},
		},
		{
			name: "not found",
			args: args{service: &MockMetricsService{
				HandleFunc: func(metric models.Metrics) error {
					return fmt.Errorf("not found")
				},
			},
				json: `{
  						"id": "LastGC",
  						"type": "gauge",
  						"value": 1744184459
					}`,
			},
			want: want{
				code: 400,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := JSONUpdateMetricHandler(context.TODO(), test.args.service, &audit.Auditor{})
			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(test.args.json))
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/update", handler)

			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func Test_GetJsonMetricHandler(t *testing.T) {
	type args struct {
		service Service
		json    string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test",
			args: args{service: &MockMetricsService{
				GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
					return &models.Metrics{
						ID:    metric.ID,
						MType: metric.MType,
						Delta: metric.Delta,
						Value: metric.Value,
						Hash:  "",
					}, nil
				},
			},
				json: `{"id": "LastGC","type": "gauge","value": 1744184459}`,
			},
			want: want{
				code:     200,
				response: `{"id": "LastGC","type": "gauge","value": 1744184459}`,
			},
		},
		{
			name: "decode error",
			args: args{service: &MockMetricsService{
				GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
					return &models.Metrics{
						ID:    metric.ID,
						MType: metric.MType,
						Delta: metric.Delta,
						Value: metric.Value,
						Hash:  "",
					}, nil
				},
			},
				json: `{"id": "LastGC","type": "gauge","value": 1744184459`,
			},
			want: want{
				code:     500,
				response: "",
			},
		},
		{
			name: "not found",
			args: args{service: &MockMetricsService{
				GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
					return nil, fmt.Errorf("метрика не найдена")
				},
			},
				json: `{"id": "LastGC","type": "gauge","value": 1744184459}`,
			},
			want: want{
				code:     404,
				response: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := GetJSONMetricHandler(context.TODO(), test.args.service)
			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBufferString(test.args.json))
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/value", handler)

			r.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if len(resBody) > 0 {
				assert.JSONEq(t, test.want.response, string(resBody))
			}
		})
	}
}
