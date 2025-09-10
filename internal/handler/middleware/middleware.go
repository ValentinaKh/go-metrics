package middleware

import (
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

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
			defer cr.Close()
		}

		ow := w
		// Проверяем, что клиент может принять сжатые данные. Если заголовок есть, сжимаем данные
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
