package services

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func detectHTMLVersion(htmlBytes []byte) string {
	tokenizer := html.NewTokenizer(bytes.NewReader(htmlBytes))

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return "Error"
		case html.DoctypeToken:
			token := tokenizer.Token()
			data := strings.ToLower(token.Data)

			switch {
			case strings.Contains(data, "html 4.01"):
				return "HTML 4.01"
			case strings.Contains(data, "xhtml"):
				return "XHTML"
			case strings.Contains(data, "html"):
				return "HTML5"
			//case strings.Contains(data, "html 3.2"):
			//	return "HTML 3.2"
			//case strings.Contains(data, "html 2.0"):
			//	return "HTML 2.0"
			default:
				return "Unknown"
			}
		default:
			return "Unknown"
		}
	}
}
