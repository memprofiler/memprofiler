package locator

import (
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// Locator stores various server subsystems
type Locator struct {
	Storage storage.Service
	Logger  *logrus.Logger
}

func newLogger(cfg *config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(cfg.Level)
	return logger
}

// NewLocator creates new Locator
func NewLocator(cfg *config.Config) (*Locator, error) {
	var l Locator

	l.Logger = newLogger(cfg.Logging)
	l.Storage = nil

	return nil, nil
}
