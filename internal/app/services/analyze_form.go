package services

import "github.com/PuerkitoBio/goquery"

func detectForm(doc *goquery.Document) bool {
	hasLoginForm := false

	doc.Find("form").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Find("input[type='password']").Length() > 0 {
			hasLoginForm = true

			return false // break early
		}

		return true
	})

	return hasLoginForm
}
