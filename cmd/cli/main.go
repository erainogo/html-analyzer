package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/erainogo/html-analyzer/internal/app/services"
	"github.com/erainogo/html-analyzer/internal/config"
	"github.com/erainogo/html-analyzer/internal/handlers"
	"github.com/erainogo/html-analyzer/pkg/constants"
)

//---------------------------------------- CLI ENTRYPOINT FOR THE APPLICATION --------------------------------------- //

type urlJob struct {
	Index int
	URL   string
}

type urlResult struct {
	Index int
	Row   []string
	Err   error
}

var logger *zap.SugaredLogger

// init ensures logger is ready before anything else runs
func init() {
	appName := fmt.Sprintf("%s-html-analyzer", *config.Config.Prefix)

	zapLogger, _ := zap.NewProduction()

	logger = zapLogger.With(zap.String("app", appName)).Sugar()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// set up http client
	hc := &http.Client{
		Timeout: 30 * time.Second,
	}

	// background routine to shut down server if signal received
	// this will wait for the ch chan to receive the exit signals from the os.
	// if received cancel the context.
	go func() {
		sig := <-ch
		logger.Infof("Got %s signal. Cancelling", sig)

		cancel()

		if tr, ok := hc.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}

		logger.Info("Server gracefully stopped")
	}()

	if len(os.Args) < constants.ARGS {
		fmt.Println("Usage: analyzer <input.csv> <output.csv>")

		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	logger.Info("Started generating report")

	inputFile, err := os.Open(inputPath)
	if err != nil {
		logger.Fatalf("Failed to open input file: %v", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	records, err := reader.ReadAll()
	if err != nil {
		logger.Fatalf("Failed to read input CSV: %v", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		logger.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write header
	err = writer.Write(constants.CsvHeader)
	if err != nil {
		logger.Fatalf("Failed to write header: %v", err)
	}

	generateCsv(ctx, logger, &records, writer, hc)

	logger.Infof("Finished analyzing. Exiting.")
	logger.Infof("Output File Generated : %s", outputPath)
}

func generateCsv(
	ctx context.Context,
	logger *zap.SugaredLogger,
	records *[][]string,
	writer *csv.Writer,
	hc *http.Client,
) {
	select {
	case <-ctx.Done():
		logger.Info("Server stopped")

		return
	default:
		if records == nil || *records == nil {
			logger.Error("records are nil")

			return
		}

		if len(*records) == 0 {
			logger.Warn("No results found")

			return
		}

		rc := *records

		service := services.NewAnalyzeService(
			ctx, hc, services.WithLogger(logger))

		cliServer := handlers.NewCliServer(
			ctx, service, handlers.CliWithLogger(logger))

		// make buffered channels for the count of the records.
		jobs := make(chan urlJob, len(rc))
		results := make(chan urlResult, len(rc))

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
		for i, record := range rc {
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
					rc[res.Index][0], "error", res.Err)

				continue
			}

			if res.Row != nil {
				if err := writer.Write(res.Row); err != nil {
					logger.Errorw("Failed to write CSV row", "url",
						rc[res.Index][0], "error", err)
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
