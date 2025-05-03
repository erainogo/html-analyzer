package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/erainogo/html-analyzer/internal/config"
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

	// background routine to shut down server if signal received
	// this will wait for the ch chan to receive the exit signals from the os.
	// if received cancel the context.
	go func() {
		sig := <-ch
		logger.Infof("Got %s signal. Cancelling", sig)

		cancel()

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

	generateCsv(ctx, logger, records, writer)

	logger.Infof("Finished analyzing. Exiting.")
	logger.Infof("Output File Generated : %s", outputPath)
}
