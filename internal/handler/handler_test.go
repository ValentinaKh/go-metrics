package handler

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockMetricsService struct {
	HandleFunc func(path string) error
}

func (m *MockMetricsService) Handle(path string) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(path)
	}
	return nil
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
				HandleFunc: func(path string) error {
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
				HandleFunc: func(path string) error {
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
