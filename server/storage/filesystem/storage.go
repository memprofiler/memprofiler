package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage"
)

var _ storage.Storage = (*defaultStorage)(nil)

// defaultStorage uses filesystem as a persistent storage;
// services - first level subdirectories;
// instances - second level subdirectories;
// sessions - third level subdirectories;
// measurements - distinct files within sessions subdirectories;
type defaultStorage struct {
	sessionStorage
	codec  codec
	cfg    *config.FilesystemStorageConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger logrus.FieldLogger
}

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

func (s *defaultStorage) NewDataSaver(desc *schema.ServiceDescription) (storage.DataSaver, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// get new sessionID for this service instance
	sessionID := s.sessionStorage.inc(desc)

	// obtain directory to store data coming from a particular service instance
	sd := &storage.SessionDescription{
		ServiceDescription: desc,
		SessionID:          sessionID,
	}
	subdirPath := s.makeSubdirPath(sd)
	if _, err := os.Stat(subdirPath); err != nil {

		if !os.IsNotExist(err) {
			return nil, err
		}

		// create directory if it doesn't exist
		if err = os.MkdirAll(subdirPath, dirPermissions); err != nil {
			return nil, fmt.Errorf("failed to create directory for service data: %v", err)
		}
	}

	return newDataSaver(subdirPath, desc, sessionID, s.cfg, &s.wg, s.codec)
}

func (s *defaultStorage) NewDataLoader(sd *storage.SessionDescription) (storage.DataLoader, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	return newDataLoader(s.makeSubdirPath(sd), sd, s.codec, s.logger, &s.wg)
}

// makeSubdirPath builds a path for a filesystem direcory with instance data
func (s *defaultStorage) makeSubdirPath(sd *storage.SessionDescription) string {
	return filepath.Join(
		s.cfg.DataDir,
		sd.ServiceDescription.GetType(),
		sd.ServiceDescription.GetInstance(),
		sd.SessionID.String(),
	)
}

func (s *defaultStorage) Quit() {
	s.cancel()
	s.wg.Wait()
}

func (s *defaultStorage) populateSessionStorage() error {

	// traverse root directory and extract subdirectories
	// in order to obtain list of services, instances, and sessions
	subdirs1, err := ioutil.ReadDir(s.cfg.DataDir)
	if err != nil {
		return err
	}

	for _, s1 := range subdirs1 {
		if s1.IsDir() {
			s1Path := filepath.Join(s.cfg.DataDir, s1.Name())
			subdirs2, err := ioutil.ReadDir(s1Path)
			if err != nil {
				return err
			}
			for _, s2 := range subdirs2 {
				desc := &schema.ServiceDescription{Type: s1.Name(), Instance: s2.Name()}

				s2Path := filepath.Join(s1Path, s2.Name())
				subdirs3, err := ioutil.ReadDir(s2Path)
				if err != nil {
					return err
				}
				for _, s3 := range subdirs3 {
					if s3.IsDir() {
						sessionID, err := storage.SessionIDFromString(s3.Name())
						if err != nil {
							return err
						}
						s.sessionStorage.register(desc, sessionID)
					}
				}
			}
		}
	}

	return nil
}

// NewStorage builds new storage that keeps measurements in separate files
func NewStorage(logger logrus.FieldLogger, cfg *config.FilesystemStorageConfig) (storage.Storage, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &defaultStorage{
		codec:          newJSONCodec(),
		sessionStorage: newSessionStorage(),
		cfg:            cfg,
		ctx:            ctx,
		cancel:         cancel,
		wg:             sync.WaitGroup{},
		logger:         logger,
	}

	// traverse dirs and find previously stored data
	if err := s.populateSessionStorage(); err != nil {
		return nil, err
	}
	return s, nil
}
