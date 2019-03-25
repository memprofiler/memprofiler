package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// BlockOnSignal is useful function for main goroutine
func BlockOnSignal(logger logrus.FieldLogger, errChan <-chan error) {

	// wait for signals or fatal errors
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case <-signalChan:
		logger.Info("Interrupt signal has been received")
	case err := <-errChan:
		if err != nil {
			logger.WithError(err).Error("Fatal error, going to terminate server")
		} else {
			logger.Warning("Going to terminate server")
		}
	}
}
