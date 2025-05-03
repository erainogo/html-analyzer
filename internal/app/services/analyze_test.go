package services

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/erainogo/html-analyzer/pkg/entities"
	"github.com/stretchr/testify/assert"
)

// Test for detecting HTML version
func TestAnalyzeHTMLVersion(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    string
	}{
		{
			name:        "HTML 4.01",
			htmlContent: strings.ToLower("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\" \"http://www.w3.org/TR/html4/strict.dtd\">"),
			expected:    "HTML 4.01",
		},
		{
			name:        "XHTML",
			htmlContent: strings.ToLower("<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">"),
			expected:    "XHTML",
		},
		{
			name:        "HTML5",
			htmlContent: strings.ToLower("<!DOCTYPE html>"),
			expected:    "HTML5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes := []byte(tt.htmlContent)
			version := detectHTMLVersion(&htmlBytes)
			assert.Equal(t, tt.expected, version)
		})
	}
}

// Test for extracting page title
func TestAnalyzeTitle(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    string
	}{
		{
			name:        "Valid Title",
			htmlContent: "<html><head><title>Test Page</title></head><body></body></html>",
			expected:    "Test Page",
		},
		{
			name:        "Empty Title",
			htmlContent: "<html><head><title></title></head><body></body></html>",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.htmlContent))
			title := doc.Find("title").Text()
			assert.Equal(t, tt.expected, title)
		})
	}
}

// Test for counting headings
func TestAnalyzeHeadings(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    map[string]int
	}{
		{
			name:        "Headings Count",
			htmlContent: "<html><body><h1>Heading 1</h1><h2>Heading 2</h2><h3>Heading 3</h3></body></html>",
			expected:    map[string]int{"h1": 1, "h2": 1, "h3": 1, "h4": 0, "h5": 0, "h6": 0},
		},
		{
			name:        "No Headings",
			htmlContent: "<html><body></body></html>",
			expected:    map[string]int{"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.htmlContent))

			headings := findHeadings(doc)

			assert.Equal(t, tt.expected, headings)
		})
	}
}

// Test for detecting login form
func TestAnalyzeLoginForm(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    bool
	}{
		{
			name:        "Login form detected",
			htmlContent: "<html><body><form><input type='password'></form></body></html>",
			expected:    true,
		},
		{
			name:        "No login form",
			htmlContent: "<html><body><form></form></body></html>",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.htmlContent))

			hasLoginForm := detectForm(doc)

			assert.Equal(t, tt.expected, hasLoginForm)
		})
	}
}

func (suite *AnalyzeTestSuite) TestParseWithUnknowHtmlVersionAndHeaders() {
	mockResult := entities.AnalysisResult{
		HTMLVersion: "Unknown",
		Title:       "",
		Headings: map[string]int{
			"h1": 1,
			"h2": 1,
			"h3": 1,
			"h4": 0,
			"h5": 0,
			"h6": 0,
		},
		Links: entities.LinkAnalysis{
			Internal:     0,
			External:     0,
			Inaccessible: 0,
		},
		HasLoginForm: false,
	}

	ctx := context.Background()

	htmlContent := "<html><body><h1>Heading 1</h1><h2>Heading 2</h2><h3>Heading 3</h3></body></html>"
	htmlBytes := []byte(htmlContent)

	result, _ := suite.service.Parse(ctx, &htmlBytes, "http://localhost/")

	suite.asserts.Equal(&mockResult, result)
}

func (suite *AnalyzeTestSuite) TestParseWithHtmlVersionAndTitle() {
	mockResult := entities.AnalysisResult{
		HTMLVersion: "HTML5",
		Title:       "Test Page",
		Headings: map[string]int{
			"h1": 0,
			"h2": 0,
			"h3": 0,
			"h4": 0,
			"h5": 0,
			"h6": 0,
		},
		Links: entities.LinkAnalysis{
			Internal:     0,
			External:     3,
			Inaccessible: 3,
		},
		HasLoginForm: false,
	}

	ctx := context.Background()

	htmlContent := "<!DOCTYPE html>\n<html>\n  <head>\n    <title>Test Page</title>\n  </head>\n  <body>\n    <a href=\"https://example.com/internal\">Internal Link</a>\n    <a href=\"https://external.com/external\">External Link</a>\n    <a href=\"https://example.com/broken\">Broken Link</a>\n  </body>\n</html>"
	htmlBytes := []byte(htmlContent)

	result, _ := suite.service.Parse(ctx, &htmlBytes, "http://localhost/")

	suite.asserts.Equal(&mockResult, result)
}

func (suite *AnalyzeTestSuite) TestParseNilHTMLBytes() {
	ctx := context.Background()

	result, err := suite.service.Parse(ctx, nil, "http://localhost/")

	suite.Error(err)
	suite.asserts.Nil(result)
	suite.asserts.EqualError(err, "html bytes nil")
}

func (suite *AnalyzeTestSuite) TestParseContextDone() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	html := []byte("<html><body><h1>Hello</h1></body></html>")

	result, err := suite.service.Parse(ctx, &html, "http://localhost/")

	suite.Error(err)
	suite.asserts.Nil(result)
	suite.asserts.Equal(context.Canceled, err)
}

func (suite *AnalyzeTestSuite) TestParseFullAnalysis() {
	mockResult := &entities.AnalysisResult{
		HTMLVersion: "HTML5",
		Title:       "Login Page",
		Headings: map[string]int{
			"h1": 1,
			"h2": 1,
			"h3": 0,
			"h4": 0,
			"h5": 0,
			"h6": 0,
		},
		Links: entities.LinkAnalysis{
			Internal:     1,
			External:     1,
			Inaccessible: 2,
		},
		HasLoginForm: true,
	}

	htmlContent := `<!DOCTYPE html>
<html>
  <head>
    <title>Login Page</title>
  </head>
  <body>
    <h1>Welcome</h1>
    <h2>Sign in below</h2>
    <form action="/login"><input type="password" /></form>
    <a href="http://localhost/internal">Internal Link</a>
    <a href="https://external.com/home">External Link</a>
    <a href="https://broken1.com/">Broken1</a>
    <a href="https://broken2.com/">Broken2</a>
  </body>
</html>`

	htmlBytes := []byte(htmlContent)
	ctx := context.Background()

	result, _ := suite.service.Parse(ctx, &htmlBytes, "http://localhost")

	suite.asserts.Equal(mockResult, result)
}
