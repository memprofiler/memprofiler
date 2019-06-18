package tsdb

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
	cfg    *config.TSDBStorageConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger logrus.FieldLogger
}

const (
	dirPermissions = 0755
)

func (s *defaultStorage) NewDataSaver(serviceDesc *schema.ServiceDescription) (storage.DataSaver, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// register new session for this service instance
	session := s.sessionStorage.registerNextSession(serviceDesc)

	// obtain directory to store data coming from a particular service instance
	subdirPath := s.makeSubdirPath(session.GetDescription())
	if _, err := os.Stat(subdirPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// create directory if it doesn't exist
		if err = os.MkdirAll(subdirPath, dirPermissions); err != nil {
			return nil, fmt.Errorf("failed to create directory for service data: %v", err)
		}
	}

	return newDataSaver(subdirPath, session.GetDescription(), s.codec, &s.wg)
}

func (s *defaultStorage) NewDataLoader(sd *schema.SessionDescription) (storage.DataLoader, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	return newDataLoader(s.makeSubdirPath(sd), sd, s.codec, s.logger, &s.wg)
}

// makeSubdirPath builds a path for a filesystem direcory with instance data
func (s *defaultStorage) makeSubdirPath(sessionDescription *schema.SessionDescription) string {
	return filepath.Join(
		s.cfg.DataDir,
		sessionDescription.GetServiceType(),
		sessionDescription.GetServiceInstance(),
		storage.SessionIDToString(sessionDescription.GetSessionId()),
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
						s.sessionStorage.registerExistingSession(
							&schema.Session{
								Description: &schema.SessionDescription{
									ServiceInstance: s2.Name(),
									ServiceType:     s1.Name(),
									SessionId:       sessionID,
								},
							})
					}
				}
			}
		}
	}

	return nil
}

// NewStorage builds new storage that keeps measurements in separate files
func NewStorage(logger logrus.FieldLogger, cfg *config.TSDBStorageConfig) (storage.Storage, error) {

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
