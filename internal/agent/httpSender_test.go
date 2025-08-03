package agent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestHTTPSender_Send_Success(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &HTTPSender{client: &http.Client{}}

	err := sender.Send(server.URL)

	assert.NoError(t, err)
}

func TestHTTPSender_Send_InvalidURL(t *testing.T) {
	sender := &HTTPSender{client: &http.Client{}}

	err := sender.Send("://invalid-url")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")
}

func TestHTTPSender_Send_NetworkError(t *testing.T) {
	client := &http.Client{
		Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return nil, &url.Error{
				Op:  "Post",
				URL: r.URL.String(),
				Err: fmt.Errorf("connection refused"),
			}
		}),
	}

	sender := &HTTPSender{client: client}

	err := sender.Send("http://localhost:12345")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}
