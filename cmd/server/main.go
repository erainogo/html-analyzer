package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/erainogo/html-analyzer/internal/app/services"
	"github.com/erainogo/html-analyzer/internal/config"
	"github.com/erainogo/html-analyzer/internal/handlers"
)

//---------------------------------------- HTTP ENTRYPOINT FOR THE APPLICATION --------------------------------------- //

// set up logger
func setUpLogger() *zap.SugaredLogger {
	appName := fmt.Sprintf("%s-html-analyzer", *config.Config.Prefix)

	zapLogger, _ := zap.NewProduction()

	return zapLogger.With(zap.String("app", appName)).Sugar()
}

func main() {
	logger := setUpLogger()

	// context for the application
	ctx, cancel := context.WithCancel(context.Background())

	// http server setup
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", *config.Config.HttpPort),
		WriteTimeout: time.Duration(*config.Config.WriteTimeOut) * time.Second,
		ReadTimeout:  time.Duration(*config.Config.ReadTimeOut) * time.Second,
	}

	// set up http client
	hc := &http.Client{
		Timeout: 30 * time.Second,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// background routine to shut down server if signal received
	// this will wait for the ch chan to receive the exit signals from the os.
	go func() {
		sig := <-ch
		logger.Infof("Got %s signal. Cancelling", sig)
		// shut down background routines
		defer cancel()

		shutdownCtx, shutdownRelease := context.WithTimeout(ctx, 5*time.Second)

		defer shutdownRelease()

		if tr, ok := hc.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Shutdown error: %s", err)
		}

		defer func() {
			if err := logger.Sync(); err != nil && !errors.Is(err, os.ErrInvalid) {
				logger.Errorf("Failed to sync logger: %v", err)
			}
		}()

		logger.Info("Server gracefully stopped")
	}()

	// service will hold the logic to get the required details from parsed url
	service := services.NewAnalyzeService(
		ctx, hc, services.WithLogger(logger))

	// http handler for routes like analyze
	srv.Handler = handlers.NewHTTPServer(
		ctx, service, handlers.WithLogger(logger))

	log.Println("Server started at :", *config.Config.HttpPort)

	// Start server
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("ListenAndServe error: %s", err)
	}
}
