package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidationPostMw(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "GET rejected",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Method GET Not Allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := ValidationPostMw(next)

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, tt.expectedStatus, w.Code)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(resBody))
		})
	}
}

func TestValidationURLRqMw(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		urlPath        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid update gauge",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/gauge/cpu/0.8",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "valid update counter",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/counter/requests/100",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "missing Content-Type",
			method:         http.MethodPost,
			contentType:    "",
			urlPath:        "/update/gauge/cpu/0.8",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Ожидался Content-Type: text/plain\n",
		},
		{
			name:           "wrong Content-Type",
			method:         http.MethodPost,
			contentType:    "application/json",
			urlPath:        "/update/gauge/cpu/0.8",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Ожидался Content-Type: text/plain\n",
		},
		{
			name:           "url has 3 parts",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/gauge/cpu",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "incorrect url\n",
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/unknown/cpu/1.0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Type unknown Not Allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := ValidationURLRqMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, tt.urlPath, nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			res := w.Result()
			defer res.Body.Close()

			if tt.expectedBody != "" {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(resBody))
			}
		})
	}
}
