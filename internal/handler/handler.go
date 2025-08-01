package handler

import (
	"github.com/ValentinaKh/go-metrics/internal/service"
	"net/http"
)

func MetricsHandler(service *service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := service.Handle(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
