package handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
)

type HealthChecker interface {
	CheckDB(ctx context.Context) error
}

func HealthHandler(ctx context.Context, h HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "text/html")
		err := h.CheckDB(timeout)
		if err != nil {
			logger.Log.Error("ping error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
