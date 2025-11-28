package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashWriter_Write(t *testing.T) {

	original := httptest.NewRecorder()
	hashW := NewHashWriter(original)

	data := []byte("metrics")
	n, err := hashW.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, hashW.body)
}

func TestHashWriter_WriteHeader(t *testing.T) {
	original := httptest.NewRecorder()
	hashW := NewHashWriter(original)

	hashW.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, hashW.statusCode)
}

func TestHashWriter_DefaultStatusCode(t *testing.T) {
	original := httptest.NewRecorder()
	hashW := NewHashWriter(original)
	assert.Equal(t, 200, hashW.statusCode)
}
