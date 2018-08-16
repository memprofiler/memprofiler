package locator

import (
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/metrics"
	"github.com/vitalyisaev2/memprofiler/server/storage"
	"github.com/vitalyisaev2/memprofiler/server/storage/filesystem"
)

// Locator stores various server subsystems
type Locator struct {
	Storage  storage.Storage
	Computer metrics.Computer
	Logger   *logrus.Logger
}

func newLogger(cfg *config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(cfg.Level)
	return logger
}

// NewLocator creates new Locator
func NewLocator(cfg *config.Config) (*Locator, error) {
	var (
		l   Locator
		err error
	)

	// 1. run logger
	l.Logger = newLogger(cfg.Logging)

	// 2. run storage
	l.Logger.Debug("Starting storage")
	if cfg.Storage.Filesystem != nil {
		l.Storage, err = filesystem.NewStorage(l.Logger, cfg.Storage.Filesystem)
	}
	if err != nil {
		return nil, err
	}

	// 3. run measurement collector
	l.Logger.Debug("Starting metrics computer")
	l.Computer = metrics.New(l.Logger)

	return &l, err
}

// Quit terminates subsystems gracefully
func (l *Locator) Quit() {
	l.Logger.Debug("Stopping storage")
	l.Storage.Quit()
	l.Logger.Debug("Stopping metrics computer")
	l.Computer.Quit()
}
