package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/metadata"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
)

var _ data.Storage = (*storage)(nil)

// storage uses filesystem as a persistent storage;
type storage struct {
	metadataStorage metadata.Storage
	codec           codec
	cfg             *config.FilesystemStorageConfig
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	logger          *zerolog.Logger
}

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

func (s *storage) NewDataSaver(instanceDesc *schema.InstanceDescription) (data.Saver, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
	}

	// register new sessionDesc for this service instance
	sessionDesc, err := s.metadataStorage.StartSession(s.ctx, instanceDesc)
	if err != nil {
		return nil, errors.Wrap(err, "start new sessionDesc")
	}

	// obtain directory to store data coming from a particular service instance
	instanceDir := s.instanceDir(sessionDesc.InstanceDescription)
	if _, err := os.Stat(instanceDir); err != nil {

		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "check directory for service instance data")
		}

		// create directory if it doesn't exist
		if err = os.MkdirAll(instanceDir, dirPermissions); err != nil {
			return nil, errors.Wrap(err, "create directory for service instance data")
		}
	}

	dataFile := s.sessionDataFile(instanceDir, sessionDesc.Id)
	s.logger.Info().Fields(map[string]interface{}{"data_file": dataFile}).Msg("Starting new session")
	s.wg.Add(1)
	return newDataSaver(dataFile, sessionDesc, s.cfg, &s.wg, s.codec, s.metadataStorage)
}

func (s *storage) NewDataLoader(sessionDesc *schema.SessionDescription) (data.Loader, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}
	filename := s.sessionDataFile(s.instanceDir(sessionDesc.InstanceDescription), sessionDesc.Id)
	return newDataLoader(filename, sessionDesc, s.codec, s.logger, &s.wg)
}

// instanceDir builds a path for a filesystem direcory with instance data
func (s *storage) instanceDir(instanceDesc *schema.InstanceDescription) string {
	return filepath.Join(
		s.cfg.DataDir,
		instanceDesc.ServiceName,
		instanceDesc.InstanceName,
	)
}

func (s *storage) sessionDataFile(dir string, sessionID int64) string {
	return filepath.Join(dir, fmt.Sprintf("%010d", sessionID))
}

func (s *storage) Quit() {
	s.cancel()
	s.wg.Wait()
}

// NewStorage builds new storage that keeps measurements in separate files
func NewStorage(
	logger *zerolog.Logger,
	cfg *config.FilesystemStorageConfig,
	metadataStorage metadata.Storage,
) (data.Storage, error) {

	// create data directory if not exists
	if _, err := os.Stat(cfg.DataDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(cfg.DataDir, dirPermissions); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &storage{
		codec:           newJSONCodec(),
		metadataStorage: metadataStorage,
		cfg:             cfg,
		ctx:             ctx,
		cancel:          cancel,
		wg:              sync.WaitGroup{},
		logger:          logger,
	}

	return s, nil
}
