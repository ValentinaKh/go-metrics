package handler

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockMetricsService struct {
	HandleFunc        func(metricType, name, value string) error
	GetMetricFunc     func(name string) (string, bool)
	GetAllMetricsFunc func() map[string]string
}

func (m *MockMetricsService) Handle(metricType, name, value string) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(metricType, name, value)
	}
	return nil
}

func (m *MockMetricsService) GetMetric(name string) (string, bool) {
	return m.GetMetricFunc(name)
}

func (m *MockMetricsService) GetAllMetrics() map[string]string {
	return m.GetAllMetricsFunc()
}

func TestMetricsHandler(t *testing.T) {
	type args struct {
		service service.Service
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
				HandleFunc: func(metricType, name, value string) error {
					return nil
				},
			},
			},
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name: "negative test",
			args: args{&MockMetricsService{
				HandleFunc: func(metricType, name, value string) error {
					return fmt.Errorf("test error")
				},
			},
			},
			want: want{
				code:     400,
				response: "test error\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := MetricsHandler(test.args.service)
			request := httptest.NewRequest(http.MethodPost, "/test", nil)

			w := httptest.NewRecorder()
			handler(w, request)

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
	type args struct {
		service service.Service
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
				GetMetricFunc: func(name string) (string, bool) {
					return "500", true
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
				GetMetricFunc: func(name string) (string, bool) {
					return "", false
				},
			},
			},
			want: want{
				code:     404,
				response: "Метрика не найдена\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := GetMetricHandler(test.args.service)
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
		service service.Service
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
			handler := GetAllMetricsHandler(test.args.service)
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
