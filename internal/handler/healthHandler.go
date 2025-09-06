package handler

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

type HealthChecker interface {
	CheckDB(ctx context.Context) error
}

func HealthHandler(ctx context.Context, h HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := h.CheckDB(ctx)
		if err != nil {
			logger.Log.Error("ping error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
