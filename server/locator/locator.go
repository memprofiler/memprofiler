package locator

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"

	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/data/filesystem"
	"github.com/memprofiler/memprofiler/server/storage/data/tsdb"
	"github.com/memprofiler/memprofiler/server/storage/metadata"
	"github.com/memprofiler/memprofiler/utils"
)

// Locator stores various server subsystems
type Locator struct {
	DataStorage     data.Storage
	MetadataStorage metadata.Storage
	Computer        metrics.Computer
	Logger          *zerolog.Logger
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

	// 2. run metadata storage
	l.Logger.Debug().Msg("Starting metadata storage")
	l.MetadataStorage, err = metadata.NewStorageSQLite(l.Logger, cfg.MetadataStorage)
	if err != nil {
		return nil, errors.Wrap(err, "metadata storage")
	}

	// 3. run data storage
	l.Logger.Debug().Msg("Starting data storage")
	switch cfg.DataStorage.Type() {
	case config.FilesystemDataStorage:
		l.DataStorage, err = filesystem.NewStorage(l.Logger, cfg.DataStorage.Filesystem, l.MetadataStorage)
	case config.TSDBDataStorage:
		l.DataStorage, err = tsdb.NewStorage(l.Logger, cfg.DataStorage.TSDB, l.MetadataStorage)
	}
	if err != nil {
		return nil, errors.Wrap(err, "data storage")
	}

	// 4. run measurement collector
	l.Logger.Debug().Msg("Starting metrics computer")
	l.Computer = metrics.NewComputer(l.Logger, l.DataStorage, cfg.Metrics)

	return &l, err
}

// Quit terminates subsystems gracefully
func (l *Locator) Quit() {
	l.Logger.Debug().Msg("Stopping storage")
	l.DataStorage.Quit()
	l.Logger.Debug().Msg("Stopping metrics computer")
	l.Computer.Quit()
}
