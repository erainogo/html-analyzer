package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/erainogo/html-analyzer/internal/core/adapters"
	"github.com/erainogo/html-analyzer/pkg/constants"
)

type CliServer struct {
	ctx     context.Context
	service adapters.AnalyzeService
	logger  *zap.SugaredLogger
}

type CliServerOption func(*CliServer)

func CliWithLogger(logger *zap.SugaredLogger) CliServerOption {
	return func(s *CliServer) {
		s.logger = logger
	}
}

func NewCliServer(ctx context.Context,
	service adapters.AnalyzeService,
	opts ...CliServerOption) adapters.CliServer {
	c := &CliServer{
		ctx:     ctx,
		service: service,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (h *CliServer) Handler(ctx context.Context, url string) (*[]string, error) {
	resp, err := getResponse(url)
	if err != nil {
		h.logger.Errorw("Failed to fetch URL: "+err.Error(), http.StatusBadGateway)

		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.Errorw("unable to close body", "error", err)
		}
	}()

	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorw("Failed to read response body", http.StatusInternalServerError)

		return nil, err
	}

	result, err := h.service.Parse(ctx, contentBytes, url)
	if err != nil {
		return nil, err
	}

	details := []string{
		url,
		result.HTMLVersion,
		result.Title,

		fmt.Sprint(result.Headings[constants.H1]),
		fmt.Sprint(result.Headings[constants.H2]),
		fmt.Sprint(result.Headings[constants.H3]),
		fmt.Sprint(result.Headings[constants.H4]),
		fmt.Sprint(result.Headings[constants.H5]),
		fmt.Sprint(result.Headings[constants.H6]),

		fmt.Sprint(result.Links.Internal),
		fmt.Sprint(result.Links.External),
		fmt.Sprint(result.Links.Inaccessible),

		fmt.Sprint(result.HasLoginForm),
	}

	return &details, nil
}
