package locator

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"

	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/filesystem"
	"github.com/memprofiler/memprofiler/server/storage/tsdb"
	"github.com/memprofiler/memprofiler/utils"
)

// Locator stores various server subsystems
type Locator struct {
	Storage  storage.Storage
	Computer metrics.Computer
	Logger   *zerolog.Logger
}

// NewLocator creates new Locator
func NewLocator(logger *zerolog.Logger, cfg *config.Config) (*Locator, error) {
	var (
		l   Locator
		err error
	)

	// 1. run logger
	l.Logger = logger

	// set global GRPC logger
	grpclog.SetLoggerV2(utils.ZeroLogToGRPCLogger(l.Logger)) // FIXME: replace to V2

	// 2. run storage
	l.Logger.Debug().Msg("Starting storage")
	switch cfg.StorageType {
	case config.StorageTypeFilesystem:
		l.Storage, err = filesystem.NewStorage(l.Logger, cfg.Filesystem)
	case config.StorageTypeTSDB:
		l.Storage, err = tsdb.NewStorage(l.Logger, cfg.TSDB)
	}
	if err != nil {
		return nil, err
	}

	// 3. run measurement collector
	l.Logger.Debug().Msg("Starting metrics computer")
	l.Computer = metrics.NewComputer(l.Logger, l.Storage, cfg.Metrics)

	return &l, err
}

// Quit terminates subsystems gracefully
func (l *Locator) Quit() {
	l.Logger.Debug().Msg("Stopping storage")
	l.Storage.Quit()
	l.Logger.Debug().Msg("Stopping metrics computer")
	l.Computer.Quit()
}
