package handlers

import (
	"net/http"

	"github.com/erainogo/html-analyzer/internal/config"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *config.Config.FEURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)

			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
