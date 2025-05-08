package handlers

import (
	"net/http"
)

// applyCorsMiddleware adds CORS headers to allow cross-origin requests.
func applyCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set common CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // for now allow all.
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}
