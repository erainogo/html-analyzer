package services

import (
	"bytes"
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"github.com/erainogo/html-analyzer/internal/core/adapters"
	"github.com/erainogo/html-analyzer/pkg/entities"
)

type AnalyzeService struct {
	logger *zap.SugaredLogger
	ctx    context.Context
}

type AnalyzeServiceOption func(*AnalyzeService)

func WithLogger(logger *zap.SugaredLogger) AnalyzeServiceOption {
	return func(u *AnalyzeService) {
		u.logger = logger
	}
}

func NewAnalyzeService(
	ctx context.Context,
	opts ...AnalyzeServiceOption,
) adapters.AnalyzeService {
	svc := &AnalyzeService{
		ctx:    ctx,
		logger: zap.NewNop().Sugar(),
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func (u AnalyzeService) Parse(ctx context.Context, htmlBytes *[]byte, url string) (*entities.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		u.logger.Info("application context done", ctx.Err())

		return nil, ctx.Err()
	default:
		if htmlBytes == nil {
			return nil, errors.New("html bytes nil")
		}
		// detect HTML version from raw HTML
		htmlVersion := detectHTMLVersion(htmlBytes)

		// parse document with goquery
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(*htmlBytes))
		if err != nil {
			return nil, errors.New("failed to parse HTML")
		}

		// retrieve the title.
		title := doc.Find("title").Text()

		// find the heading count
		headings := findHeadings(doc)

		baseHost := getHost(url)
		// concurrently checking to improve the look-up
		linkResult := analyzeLinks(ctx, doc, baseHost, u.logger)

		// Login form detection
		// going to use password keyword for the look-up
		// usually page yields a small number of forms
		hasLoginForm := detectForm(doc)

		// consolidate all the results for the response.
		return &entities.AnalysisResult{
			HTMLVersion: htmlVersion,
			Title:       title,
			Headings:    headings,
			Links: entities.LinkAnalysis{
				Internal:     linkResult.Internal,
				External:     linkResult.External,
				Inaccessible: linkResult.Inaccessible,
			},
			HasLoginForm: hasLoginForm,
		}, nil
	}
}
