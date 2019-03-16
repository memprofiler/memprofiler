package locator

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"

	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/filesystem"
	"github.com/memprofiler/memprofiler/utils"
)

// Locator stores various server subsystems
type Locator struct {
	Storage  storage.Storage
	Computer metrics.Computer
	Logger   logrus.FieldLogger
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

	// set global GRPC logger
	grpclog.SetLoggerV2(utils.LogrusToGRPCLogger(l.Logger)) // FIXME: replace to V2

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
	l.Computer = metrics.NewComputer(l.Logger, l.Storage, cfg.Metrics)

	return &l, err
}

// Quit terminates subsystems gracefully
func (l *Locator) Quit() {
	l.Logger.Debug("Stopping storage")
	l.Storage.Quit()
	l.Logger.Debug("Stopping metrics computer")
	l.Computer.Quit()
}
