package services

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"github.com/erainogo/html-analyzer/pkg/constants"
)

type LinkStats struct {
	Internal     int
	External     int
	Inaccessible int
}

type linkCheckResult struct {
	isInternal   bool
	isAccessible bool
}

type linkJob struct {
	href string
}

// analyzeLinks this is the most time-consuming task in whole request.
// because it's Network I/O Bound: each HEAD request goes out to the internet
// so we can use worker pool concurrency pattern to check the status of the links
// concurrent execution, we can improve the performance of the request
func analyzeLinks(ctx context.Context, doc *goquery.Document, baseHost string, logger *zap.SugaredLogger) LinkStats {
	jobs := make(chan linkJob)
	results := make(chan linkCheckResult)

	var wg sync.WaitGroup

	// Start fixed number of workers to check
	for i := 0; i < constants.WorkerCount; i++ {
		wg.Add(1)
		workerID := i

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						logger.Infof("channel closed, exiting analyzeLinks worker %v", workerID)

						return
					}

					href := job.href

					isFullURL := strings.HasPrefix(href, "http")

					isInternal := isInternalLink(href, baseHost)

					accessible := true
					if isFullURL {
						accessible = isLinkAccessible(href)
					}

					result := linkCheckResult{
						isInternal:   isInternal,
						isAccessible: accessible,
					}

					select {
					case <-ctx.Done():
						return
					case results <- result:
					}
				}
			}
		}()
	}

	// Feed jobs to workers
	go func() {
		defer close(jobs)

		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				jobs <- linkJob{href: href}
			}
		})
	}()

	// Close results when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and update stats
	stats := LinkStats{}

	for res := range results {
		if res.isInternal {
			stats.Internal++
		} else {
			stats.External++
		}

		if !res.isAccessible {
			stats.Inaccessible++
		}
	}

	return stats
}

func isLinkAccessible(link string) bool {
	resp, err := http.Head(link)
	if err != nil || resp.StatusCode >= 400 {
		return false
	}

	return true
}

func isInternalLink(href, baseHost string) bool {
	parsed, err := url.Parse(href)
	if err != nil {
		return false
	}

	if parsed.Host == "" {
		return true
	}

	return parsed.Host == baseHost
}
