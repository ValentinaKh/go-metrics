package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"github.com/ValentinaKh/go-metrics/internal/utils"
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

func TestNewPostSender(t *testing.T) {
	host := "localhost:8080"
	secureKey := "test-key"

	retrier := retry.NewRetrier(
		retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), 3),
		retry.NewStaticDelayStrategy([]time.Duration{10 * time.Millisecond}),
		&retry.SleepTimeProvider{},
	)

	sender := NewPostSender(host, retrier, secureKey, nil)

	require.NotNil(t, sender)

	assert.NotNil(t, sender.client)
	assert.Equal(t, "http://localhost:8080/updates/", sender.url)
	assert.Equal(t, secureKey, sender.secureKey)
	assert.Same(t, retrier, sender.retrier)
}

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
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(r.Body)

		gz, err := gzip.NewReader(bytes.NewReader(compressedBody))
		require.NoError(t, err)
		defer func(gz *gzip.Reader) {
			err := gz.Close()
			if err != nil {
				panic(err)
			}
		}(gz)

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

func TestHTTPSender_Send_WithSecureKey(t *testing.T) {

	secureKey := "secret-key"
	rq := []byte(`{"id":"test","type":"gauge","value":42.0}`)

	var compressedBody bytes.Buffer
	gz := gzip.NewWriter(&compressedBody)
	_, err := gz.Write(rq)
	if err != nil {
		t.Fatal(err)
	}
	err = gz.Close()
	if err != nil {
		t.Fatal(err)
	}
	expectedHash := utils.Hash(secureKey, compressedBody.Bytes())
	expectedHashStr := hex.EncodeToString(expectedHash)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, receivedReq *http.Request) {
		require.NotNil(t, receivedReq)
		assert.Equal(t, "application/json", receivedReq.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", receivedReq.Header.Get("Content-Encoding"))

		assert.Equal(t, expectedHashStr, receivedReq.Header.Get("HashSHA256"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &HTTPSender{client: resty.New(), url: server.URL, retrier: retry.NewRetrier(
		retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), 1),
		retry.NewStaticDelayStrategy([]time.Duration{1}),
		&retry.SleepTimeProvider{}), secureKey: secureKey}

	err = sender.Send(rq)
	require.NoError(t, err)
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

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "host with port",
			host:     "localhost:8080",
			expected: "http://localhost:8080/updates/",
		},
		{
			name:     "Domain",
			host:     "example.com",
			expected: "http://example.com/updates/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildURL(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}
