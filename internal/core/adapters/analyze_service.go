package adapters

import (
	"context"
	"github.com/erainogo/html-analyzer/pkg/entities"
)

type AnalyzeService interface {
	Parse(context.Context, *[]byte, string) (*entities.AnalysisResult, error)
}
