package rest

import (
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/audit"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditHandler_Update_Success(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "Alloc", MType: "gauge", Value: float64Ptr(123456.0)},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, receivedReq *http.Request) {

		require.NotNil(t, receivedReq)
		assert.Equal(t, "application/json", receivedReq.Header.Get("Content-Type"))

		body, err := io.ReadAll(receivedReq.Body)
		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Logf("Error closing body: %v", err)
			}
		}(receivedReq.Body)

		var dto = audit.Dto{}
		err = json.Unmarshal(body, &dto)
		require.NoError(t, err)

		assert.Equal(t, "localhost", dto.IPAddress)
		assert.Equal(t, metrics, dto.Metrics)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	handler := NewAuditHandler(ts.URL)

	handler.Update(metrics, "localhost")

}

func float64Ptr(v float64) *float64 { return &v }
