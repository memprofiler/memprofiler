package tsdb

import (
	"context"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/data/tsdb/prometheus"
	"github.com/memprofiler/memprofiler/server/storage/metadata"

	"github.com/rs/zerolog"
)

var _ data.Storage = (*storage)(nil)

// storage uses Prometheus TSDB as a persistent storage
type storage struct {
	codec           codec
	cfg             *config.TSDBStorageConfig
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	logger          *zerolog.Logger
	tsdbStorage     prometheus.TSDB
	metadataStorage metadata.Storage
}

func (s *storage) NewDataSaver(instanceDesc *schema.InstanceDescription) (data.Saver, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// register new session for this service instance
	sessionDesc, err := s.metadataStorage.StartSession(s.ctx, instanceDesc)
	if err != nil {
		return nil, errors.Wrap(err, "data session")
	}
	return newDataSaver(sessionDesc, s.codec, &s.wg, s.tsdbStorage, s.metadataStorage)
}

func (s *storage) NewDataLoader(sd *schema.SessionDescription) (data.Loader, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	return newDataLoader(sd, s.codec, s.logger, &s.wg, s.tsdbStorage)
}

func (s *storage) Quit() {
	s.cancel()
	s.wg.Wait()
}

// NewStorage builds new storage that keeps measurements in tsdb
func NewStorage(
	logger *zerolog.Logger,
	cfg *config.TSDBStorageConfig,
	metadataStorage metadata.Storage,
) (data.Storage, error) {
	goKitWrapper, err := NewGoKitLogWrapper(logger)
	if err != nil {
		return nil, err
	}

	// create data directory if not exists
	if _, err = os.Stat(cfg.DataDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(cfg.DataDir, 0750); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// create storage
	stor, err := prometheus.OpenTSDB(cfg.DataDir, goKitWrapper)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &storage{
		codec:           newB64Codec(),
		metadataStorage: metadataStorage,
		cfg:             cfg,
		ctx:             ctx,
		cancel:          cancel,
		wg:              sync.WaitGroup{},
		logger:          logger,
		tsdbStorage:     stor,
	}
	return s, nil
}
