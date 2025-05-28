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
func analyzeLinks(
	ctx context.Context,
	hc *http.Client,
	doc *goquery.Document,
	baseHost string,
	logger *zap.SugaredLogger,
) LinkStats {
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

					accessible := false
					if isFullURL {
						accessible = isLinkAccessible(ctx, href, hc)
					}

					// exclude non navigational links early.
					if filterNonNavigationalLinks(href) {
						continue
					}

					isInternal := isInternalLink(href, baseHost)

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

func isLinkAccessible(ctx context.Context, link string, hc *http.Client) bool {
	// ctx added to avoid request hanging
	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	if err != nil {
		return false
	}
	// some servers might block or rate limit, lets use the user agent for minimize that
	req.Header.Set("User-Agent", constants.USERAGENT)
	// asks the server for just the headers, not the entire response body
	//this is much faster and cheaper
	resp, err := hc.Do(req)

	defer func() {
		if resp != nil {
			err := resp.Body.Close()
			if err != nil {
				return
			}
		}
	}()

	// some servers don’t support head
	if err != nil || resp.StatusCode >= constants.UNAUTHORIZEDCODE {
		req.Method = "GET"
		// download the whole response using GET
		resp, err = hc.Do(req)
		if err != nil || resp.StatusCode >= constants.UNAUTHORIZEDCODE {
			return false
		}
	}

	return true
}

func isInternalLink(href, baseHost string) bool {
	parsed, err := url.Parse(href)
	if err != nil {
		return false
	}

	// if Host is empty, these are internal by nature
	// because hey reference the same domain
	if parsed.Host == "" {
		return true
	}

	// removes the "www."
	normalize := func(host string) string {
		return strings.TrimPrefix(strings.ToLower(host), "www.")
	}

	// compare parsed host with base host
	return normalize(parsed.Host) == normalize(baseHost)
}

func filterNonNavigationalLinks(href string) bool {
	return strings.HasPrefix(href, "javascript:") ||
		strings.HasPrefix(href, "#") ||
		strings.HasPrefix(href, "mailto:") ||
		strings.HasPrefix(href, "tel:") ||
		strings.HasPrefix(href, "data:") ||
		strings.HasPrefix(href, "blob:") ||
		strings.HasPrefix(href, "about:") ||
		strings.HasPrefix(href, "file:") ||
		strings.HasPrefix(href, "chrome:") ||
		strings.HasPrefix(href, "edge:")
}
