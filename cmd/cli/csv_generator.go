package main

import (
	"context"
	"encoding/csv"
	"sync"

	"go.uber.org/zap"

	"github.com/erainogo/html-analyzer/internal/app/services"
	"github.com/erainogo/html-analyzer/internal/handlers"
	"github.com/erainogo/html-analyzer/pkg/constants"
)

func generateCsv(
	ctx context.Context,
	logger *zap.SugaredLogger,
	records [][]string,
	writer *csv.Writer,
) {
	select {
	case <-ctx.Done():
		logger.Info("Server stopped")

		return
	default:
		service := services.NewAnalyzeService(
			ctx, services.WithLogger(logger))

		cliServer := handlers.NewCliServer(
			ctx, service, handlers.CliWithLogger(logger))

		// make buffered channels for the count of the records.
		jobs := make(chan urlJob, len(records))
		results := make(chan urlResult, len(records))

		var wg sync.WaitGroup

		// Start worker pool
		for i := 0; i < constants.CLIWorkerCount; i++ {
			wg.Add(1)
			workerID := i

			go func() {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						logger.Infof("cli worker id %v exiting due to cancellation", workerID)

						return
					case job, ok := <-jobs:
						if !ok {
							logger.Infof("cli worker id %v exiting channel closed", workerID)

							return
						}

						row, err := cliServer.Handler(ctx, job.URL)

						logger.Infof("processed row %v", row)

						results <- urlResult{
							Index: job.Index,
							Row:   safeRow(row),
							Err:   err,
						}
					}
				}
			}()
		}

		// Send jobs
		for i, record := range records {
			if len(record) == 0 {
				continue
			}

			jobs <- urlJob{Index: i, URL: record[0]}
		}

		close(jobs)

		// wait for collect results
		go func() {
			wg.Wait()
			close(results)
		}()

		for res := range results {
			if res.Err != nil {
				logger.Errorw("Error analyzing URL", "url",
					records[res.Index][0], "error", res.Err)

				continue
			}

			if res.Row != nil {
				if err := writer.Write(res.Row); err != nil {
					logger.Errorw("Failed to write CSV row", "url",
						records[res.Index][0], "error", err)
				}
			}
		}
	}
}

func safeRow(row *[]string) []string {
	if row == nil {
		return nil
	}
	return *row
}
