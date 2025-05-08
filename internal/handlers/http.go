package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

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

	return applyCorsMiddleware(h)
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
			h.logger.Errorw("failed to decode request body", "error", err)

			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)

			return
		}

		if body.URL == "" {
			h.logger.Warn("empty URL provided")

			http.Error(w, "URL must not be empty", http.StatusBadRequest)

			return
		}

		parsedURL, err := url.ParseRequestURI(body.URL)
		if err != nil {
			h.logger.Errorw("invalid URL", "url", body.URL, "error", err)

			http.Error(w, "Invalid URL format", http.StatusBadRequest)

			return
		}

		// Safe HTTP request with timeout
		reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			h.logger.Errorw("failed to create request", "error", err)

			http.Error(w, "Failed to prepare request", http.StatusInternalServerError)

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			h.logger.Errorw("failed to fetch URL", "url", parsedURL.String(), "error", err)

			http.Error(w, "Failed to fetch the provided URL", http.StatusBadGateway)

			return
		}
		defer resp.Body.Close()

		contentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			h.logger.Errorw("failed to read response body", "error", err)

			http.Error(w, "Failed to read response", http.StatusInternalServerError)

			return
		}

		// call with both HTML content and URL
		result, err := h.service.Parse(ctx, &contentBytes, body.URL)
		if err != nil {
			h.logger.Errorw("parsing failed", "url", body.URL, "error", err)

			http.Error(w, "Failed to analyze content", http.StatusBadGateway)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result); err != nil {
			h.logger.Errorw("failed to encode result", "error", err)

			http.Error(w, "Failed to encode response", http.StatusInternalServerError)

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
	h.serve(w, r)
}

func (h *HttpServer) serve(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	h.mux.ServeHTTP(w, r)

	duration := time.Since(startTime)

	h.logger.Infof("Completed request: method=%s, url=%s, duration=%v",
		r.Method, r.URL.String(), duration)
}
