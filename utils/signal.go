package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
)

// BlockOnSignal is useful function for main goroutine
func BlockOnSignal(logger *zerolog.Logger, errChan <-chan error) {

	// wait for signals or fatal errors
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case <-signalChan:
		logger.Info().Msg("Interrupt signal has been received")
	case err := <-errChan:
		if err != nil {
			logger.Err(err).Msg("Fatal error, going to terminate server")
		} else {
			logger.Warn().Msg("Going to terminate server")
		}
	}
}
