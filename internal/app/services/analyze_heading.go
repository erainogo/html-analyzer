package services

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/erainogo/html-analyzer/pkg/constants"
)

func findHeadings(doc *goquery.Document) map[string]int {
	headings := map[string]int{}

	for i := 1; i <= constants.HeaderCount; i++ {
		tag := fmt.Sprintf("h%d", i)
		headings[tag] = doc.Find(tag).Length()
	}

	return headings
}
