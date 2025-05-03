package handlers

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/erainogo/html-analyzer/internal/core/adapters"
	"github.com/erainogo/html-analyzer/pkg/entities"
)

type HttpServer struct {
	mux     *http.ServeMux
	service adapters.AnalyzeService
	ctx     context.Context
	logger  *zap.SugaredLogger
}

type HttpServerOption func(*HttpServer)

func WithLogger(logger *zap.SugaredLogger) HttpServerOption {
	return func(s *HttpServer) {
		s.logger = logger
	}
}

func NewHTTPServer(
	ctx context.Context,
	service adapters.AnalyzeService,
	opts ...HttpServerOption) http.Handler {
	h := &HttpServer{
		ctx:     ctx,
		mux:     http.NewServeMux(),
		service: service,
	}

	for _, opt := range opts {
		opt(h)
	}

	h.registerRoutes(ctx)

	return h
}

// registerRoutes register the http routes
func (h *HttpServer) registerRoutes(ctx context.Context) {
	h.mux.HandleFunc("/analyze", h.analyzeHandler(ctx))
	h.mux.HandleFunc("/health", h.HealthHandler())
}

// analyzeHandler handler for the /analyze route.
func (h *HttpServer) analyzeHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

			return
		}

		var body entities.RequestBody

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)

			return
		}

		resp, err := getResponse(body.URL)
		if err != nil {
			http.Error(w, "Failed to fetch URL: "+err.Error(), http.StatusBadGateway)

			return
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				h.logger.Errorw("unable to close body", "error", err)
			}
		}()

		contentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)

			return
		}

		// Call with both HTML content and URL
		result, err := h.service.Parse(ctx, &contentBytes, body.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "Failed to encode result", http.StatusInternalServerError)

			return
		}
	}
}

// HealthHandler handler for the /health route.
func (h *HttpServer) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			return
		}
	}
}

func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
