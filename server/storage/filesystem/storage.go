package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
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
	logger *logrus.Logger
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
	subdirPath := s.makeSubdirPath(desc, sessionID)
	if _, err := os.Stat(subdirPath); err != nil {

		if !os.IsNotExist(err) {
			return nil, err
		}

		// create directory if it doesn't exist
		if err = os.MkdirAll(subdirPath, dirPermissions); err != nil {
			return nil, fmt.Errorf("failed to create directory for service data: %v", err)
		}
	}

	saver := &defaultDataSaver{
		subdirPath: subdirPath,
		sessionID:  sessionID,
		codec:      s.codec,
		cfg:        s.cfg,
		wg:         &s.wg,
	}

	return saver, nil
}

func (s *defaultStorage) NewDataLoader(
	desc *schema.ServiceDescription,
	sessionID storage.SessionID,
) (storage.DataLoader, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// prepare list of files with measurement dumps
	subdirPath := s.makeSubdirPath(desc, sessionID)

	loader := &defaultDataLoader{
		subdirPath: subdirPath,
		codec:      s.codec,
		sessionID:  sessionID,
		wg:         &s.wg,
	}
	return loader, nil
}

// makeSubdirPath builds a path for a filesystem direcory with instance data
func (s *defaultStorage) makeSubdirPath(
	desc *schema.ServiceDescription,
	sessionID storage.SessionID,
) string {
	return filepath.Join(
		s.cfg.DataDir,
		desc.GetType(),
		desc.GetInstance(),
		sessionID.String(),
	)
}

const timeFormat = "%d%02d%02d-%02d%02d%02d"

// makeFilePath creates a path for a file to store a distinct measurement
func makeFilePath(subdirPath string, tstamp *timestamp.Timestamp) (string, error) {

	t, err := ptypes.Timestamp(tstamp)
	if err != nil {
		return "", err
	}

	dump := fmt.Sprintf(timeFormat,
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return filepath.Join(subdirPath, dump), nil
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
func NewStorage(logger *logrus.Logger, cfg *config.FilesystemStorageConfig) (storage.Storage, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &defaultStorage{
		codec:          &jsonCodec{},
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
