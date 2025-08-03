package agent

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSender_Send_Success(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &HTTPSender{client: resty.New()}

	err := sender.Send(server.URL)

	assert.NoError(t, err)
}

func TestHTTPSender_Send_InvalidURL(t *testing.T) {
	sender := &HTTPSender{client: resty.New()}

	err := sender.Send("://invalid-url")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")
}
