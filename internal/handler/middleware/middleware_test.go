package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/ValentinaKh/go-metrics/internal/crypto"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				_, err := w.Write([]byte("OK"))
				if err != nil {
					return
				}
			})

			handler := ValidationPostMw(next)

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, tt.expectedStatus, w.Code)

			defer func(r *http.Response) {
				err := r.Body.Close()
				if err != nil {
					panic(err)
				}
			}(res)
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
			defer func(r *http.Response) {
				err := r.Body.Close()
				if err != nil {
					panic(err)
				}
			}(res)

			if tt.expectedBody != "" {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(resBody))
			}
		})
	}
}

func TestGzipMW(t *testing.T) {
	requestBody := `{
  						"id": "LastGC",
  						"type": "gauge",
  						"value": 1744184459
					}`

	successBody := `{
  						"id": "LastGC",
  						"type": "gauge",
  						"value": 1744184459
					}`

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(successBody))
		if err != nil {
			return
		}
	})

	handler := GzipMW(next)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func(r *http.Response) {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func(r *http.Response) {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp)

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}

func TestDecryptMW_Success(t *testing.T) {

	pubPath, privatePath := createTestKeys(t)

	public, err := crypto.NewPublicKeyService(pubPath)
	require.NoError(t, err)

	private, err := crypto.NewPrivateKeyService(privatePath)
	require.NoError(t, err)

	originalMsg := []byte(`{
  						"id": "LastGC",
  						"type": "gauge",
  						"value": 1744184459
					}`)
	encrypted, err := public.Transform(originalMsg)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	var receivedBody []byte
	orHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.WriteHeader(http.StatusOK)
	})

	handler := DecryptMW(private)
	wh := handler(orHandler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(encrypted))
	w := httptest.NewRecorder()

	wh.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, originalMsg, receivedBody)
}

func TestDecryptMW_DecryptError(t *testing.T) {
	w := httptest.NewRecorder()
	_, privatePath := createTestKeys(t)

	private, err := crypto.NewPrivateKeyService(privatePath)
	require.NoError(t, err)

	handler := DecryptMW(private)

	wh := handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Next handler should not be called")
	}))

	wh.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("test"))))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func createTestKeys(t *testing.T) (publicPath, privatePath string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privBytes := pem.EncodeToMemory(privBlock)

	template := &x509.Certificate{
		SerialNumber: randomSerial(),
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create self-signed cert: %v", err)
	}
	pubBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}
	pubBytes := pem.EncodeToMemory(pubBlock)

	tempDir := t.TempDir()
	publicPath = filepath.Join(tempDir, "public.pem")
	privatePath = filepath.Join(tempDir, "private.pem")

	if err := os.WriteFile(publicPath, pubBytes, 0600); err != nil {
		t.Fatalf("Failed to write public key: %v", err)
	}
	if err := os.WriteFile(privatePath, privBytes, 0600); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	return publicPath, privatePath
}

func randomSerial() *big.Int {
	serial, _ := rand.Int(rand.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil))
	return serial
}
