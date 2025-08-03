package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etoneja/go-gophermart/internal/api"
	"github.com/etoneja/go-gophermart/internal/processor"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	apiLogger := baseLogger.With().Str("component", "api").Logger()
	processorLogger := baseLogger.With().Str("component", "processor").Logger()

	application, err := app.NewAPIApp(ctx, apiLogger)
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("Failed to initialize application")
	}

	processorCtx, processorCancel := context.WithCancel(ctx)
	defer processorCancel()

	processor := processor.NewOrderProcessor(application.Config, application.DB, processorLogger)
	go processor.Run(processorCtx)

	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- application.Run()
	}()

	select {
	case sig := <-sigChan:
		baseLogger.Info().Msgf("Received signal %v, initiating shutdown...", sig)
	case err := <-serverErrChan:
		if err != nil {
			baseLogger.Error().Err(err).Msg("Server error, initiating shutdown...")
		}
	}

	baseLogger.Info().Msg("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		baseLogger.Error().Err(err).Msg("Error during server shutdown")
	}

	processorCancel()

	processor.Stop()

	baseLogger.Info().Msg("Server stopped gracefully")
}
