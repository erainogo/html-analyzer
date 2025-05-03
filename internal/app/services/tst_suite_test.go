package services

import (
	"context"
	"testing"

	"github.com/erainogo/html-analyzer/internal/core/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AnalyzeTestSuite struct {
	suite.Suite
	asserts *assert.Assertions
	service adapters.AnalyzeService
}

func (suite *AnalyzeTestSuite) SetupTest() {
	suite.asserts = assert.New(suite.T())
	ctx := context.Background()

	suite.service = NewAnalyzeService(ctx)
}

func TestAnalyzeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AnalyzeTestSuite))
}
