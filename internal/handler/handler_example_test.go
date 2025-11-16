package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/audit"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
)

func ExampleMetricsHandler() {
	h := MetricsHandler(context.TODO(), &MockMetricsService{
		HandleFunc: func(metric models.Metrics) error {
			return nil
		},
	})
	request := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu_load/0.85", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h)

	r.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response Status: %s, Response body %s", res.Status, string(resBody))
}
func ExampleGetMetricHandler() {
	var res int64 = 500
	handler := GetMetricHandler(context.TODO(), &MockMetricsService{
		GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
			return &models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &res,
				Value: metric.Value,
				Hash:  "",
			}, nil
		},
	})
	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", handler)
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/value/counter/custom")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body := resp.Body
	data, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response Status: %s, Response body %s", resp.Status, string(data))
}

func ExampleGetAllMetricsHandler() {
	handler := GetAllMetricsHandler(context.TODO(), &MockMetricsService{
		GetAllMetricsFunc: func() map[string]string {
			return map[string]string{"cpu": "0.54"}
		},
	})
	r := chi.NewRouter()
	r.Get("/", handler)
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body := resp.Body
	data, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response Status: %s, Response body %s", resp.Status, string(data))
}

func ExampleJSONUpdateMetricHandler() {
	handler := JSONUpdateMetricHandler(context.TODO(), &MockMetricsService{
		HandleFunc: func(metric models.Metrics) error {
			return nil
		},
	}, &audit.Auditor{})
	request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{"id": "LastGC","type": "gauge","value": 1744184459}`))
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/update", handler)

	r.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()

	fmt.Printf("Response Status: %s", res.Status)
}

func ExampleGetJSONMetricHandler() {
	handler := GetJSONMetricHandler(context.TODO(), &MockMetricsService{
		GetMetricFunc: func(metric models.Metrics) (*models.Metrics, error) {
			return &models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: metric.Delta,
				Value: metric.Value,
				Hash:  "",
			}, nil
		},
	})
	request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBufferString(`{"id": "LastGC","type": "gauge","value": 1744184459}`))
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/value", handler)

	r.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response Status: %s, Response body %s", res.Status, string(resBody))
}
