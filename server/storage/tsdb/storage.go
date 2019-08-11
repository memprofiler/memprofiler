package tsdb

import (
	"context"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/tsdb/prometheus"
)

var _ storage.Storage = (*tsdbStorage)(nil)

// tsdbStorage uses prometheus tsdb as a persistent storage
type tsdbStorage struct {
	storage.SessionStorage
	codec  codec
	cfg    *config.TSDBStorageConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger logrus.FieldLogger
	stor   prometheus.TSDB
}

func (s *tsdbStorage) NewDataSaver(serviceDesc *schema.ServiceDescription) (storage.DataSaver, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// register new session for this service instance
	session := s.SessionStorage.RegisterNextSession(serviceDesc)
	return newDataSaver(session.GetDescription(), s.codec, &s.wg, s.stor)
}

func (s *tsdbStorage) NewDataLoader(sd *schema.SessionDescription) (storage.DataLoader, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	return newDataLoader(sd, s.codec, s.logger, &s.wg, s.stor)
}

func (s *tsdbStorage) Quit() {
	s.cancel()
	s.wg.Wait()
}

// NewStorage builds new storage that keeps measurements in tsdb
func NewStorage(logger logrus.FieldLogger, cfg *config.TSDBStorageConfig) (storage.Storage, error) {
	// TODO: wrap to logrus interface
	var (
		writer  = log.NewSyncWriter(os.Stdout)
		logger2 = log.NewLogfmtLogger(writer)
	)

	// create data directory if not exists
	if _, err := os.Stat(cfg.DataDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// create storage
	stor, err := prometheus.OpenTSDB(cfg.DataDir, logger2)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &tsdbStorage{
		codec:          newB64Codec(),
		SessionStorage: storage.NewSessionStorage(),
		cfg:            cfg,
		ctx:            ctx,
		cancel:         cancel,
		wg:             sync.WaitGroup{},
		logger:         logger,
		stor:           stor,
	}
	return s, nil
}
