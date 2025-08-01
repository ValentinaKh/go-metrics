package middleware

import (
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"net/http"
)

func ValidationPostMw(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {

			http.Error(w, "Method "+r.Method+" Not Allowed", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ValidationUrlRqMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(w, "Ожидался Content-Type: text/plain", http.StatusBadRequest)
			return
		}
		matches := utils.ParseUrl(r.URL.Path)

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
