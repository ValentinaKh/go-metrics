package agent

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/retry"
)

func TestHTTPSender_Send_Success(t *testing.T) {
	expected := `{
					"id": "LastGC",
  					"type": "gauge",
  					"value": 1744184459
				}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		compressedBody, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer r.Body.Close()

		gz, err := gzip.NewReader(bytes.NewReader(compressedBody))
		require.NoError(t, err)
		defer gz.Close()

		uncompressedBody, err := io.ReadAll(gz)
		require.NoError(t, err)

		assert.JSONEq(t, expected, string(uncompressedBody))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &HTTPSender{client: resty.New(), url: server.URL, retrier: retry.NewRetrier(
		retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), 1),
		retry.NewStaticDelayStrategy([]time.Duration{1}),
		&retry.SleepTimeProvider{})}

	err := sender.Send([]byte(expected))

	assert.NoError(t, err)
}

func TestHTTPSender_Send_InvalidURL(t *testing.T) {
	sender := &HTTPSender{client: resty.New(), url: "://invalid-url", retrier: retry.NewRetrier(
		retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), 1),
		retry.NewStaticDelayStrategy([]time.Duration{1}),
		&retry.SleepTimeProvider{})}

	err := sender.Send([]byte(`{}`))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")
}
