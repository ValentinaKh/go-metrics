package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/crypto"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/utils"
)

const hashHeader = "HashSHA256"

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// ValidationPostMw deprecated
func ValidationPostMw(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {

			http.Error(w, "Method "+r.Method+" Not Allowed", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ValidationURLRqMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matches := utils.ParseURL(r.URL.Path)

		if len(matches) != 4 {
			http.Error(w, "incorrect url", http.StatusNotFound)
			return
		}
		if matches[1] != models.Counter && matches[1] != models.Gauge {
			http.Error(w, "Type "+matches[1]+" Not Allowed", http.StatusBadRequest)
			return
		}
		if matches[2] == "" {
			http.Error(w, "Name not found", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggingMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData: &responseData{
				status: 0,
				size:   0,
			},
		}
		defer func() {
			logger.Log.Info("HTTP запрос завершён",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", lw.responseData.status),
				zap.Int64("duration", time.Since(start).Milliseconds()),
				zap.Int("size", lw.responseData.size),
			)
		}()

		next.ServeHTTP(&lw, r)
	})
}

func GzipMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Проверяем в каком виде клиент прислал данные
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func(cr *compressReader) {
				err := cr.Close()
				if err != nil {
					logger.Log.Error("Error closing compress reader", zap.Error(err))
				}
			}(cr)
		}

		ow := w
		// Проверяем, что клиент может принять сжатые данные. Если заголовок есть, сжимаем данные
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer func(cw *compressWriter) {
				err := cw.Close()
				if err != nil {
					logger.Log.Error("Error closing compress writer", zap.Error(err))
				}
			}(cw)
		}

		next.ServeHTTP(ow, r)
	})
}

func ValidateHashMW(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerHash := r.Header.Get(hashHeader)
			if secretKey == "" || headerHash == "" {
				next.ServeHTTP(w, r)
				return
			}
			hash, err := hex.DecodeString(headerHash)
			if err != nil {
				logger.Log.Error("Error decoding hash", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Failed to read request body")
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			err = r.Body.Close()
			if err != nil {
				return
			}

			requestHash := utils.Hash(secretKey, body)
			if !hmac.Equal(hash, requestHash) {
				logger.Log.Error("not expected hash", zap.String("requestHash", fmt.Sprintf("%x", requestHash)),
					zap.String("headerHash", fmt.Sprintf("%x", hash)))
				http.Error(w, "Invalid request hash", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			next.ServeHTTP(w, r)
		})
	}
}

func HashResponseMW(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secretKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			hrw := NewHashWriter(w)

			next.ServeHTTP(hrw, r)

			if hrw.statusCode != 0 {
				w.WriteHeader(hrw.statusCode)
			}
			if len(hrw.body) > 0 {
				hash := utils.Hash(secretKey, hrw.body)
				w.Header().Set(hashHeader, string(hash))
				_, err := w.Write(hrw.body)
				if err != nil {
					return
				}
			}
		})
	}
}

func DecryptMW(cs *crypto.CryptoService[*rsa.PrivateKey, *rsa.PrivateKey]) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Failed to read request body", zap.Error(err))
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			err = r.Body.Close()
			if err != nil {
				http.Error(w, "Failed to close body", http.StatusBadRequest)
				return
			}

			var decrypted = body
			if cs != nil {
				decrypted, err = cs.Transform(body)
				if err != nil {
					logger.Log.Error("Failed to decrypt body", zap.Error(err))
					http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
					return
				}
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))

			next.ServeHTTP(w, r)
		})
	}
}
