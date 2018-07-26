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
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ storage.Storage = (*defaultStorage)(nil)

// defaultStorage uses filesystem as a persistent storage
type defaultStorage struct {
	codec codec
	cfg   *config.FilesystemStorageConfig

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
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

	// obtain directory to store data coming from a particular service instance
	subdirPath := s.makeSubdirPath(desc)
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
		codec:      s.codec,
		sessionID:  "", // FIXME: fill it
		cfg:        s.cfg,
		wg:         &s.wg,
	}

	return saver, nil
}

func (s *defaultStorage) NewDataLoader(desc *schema.ServiceDescription) (storage.DataLoader, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
	}

	// prepare list of files with measurement dumps
	subdirPath := s.makeSubdirPath(desc)

	loader := &defaultDataLoader{
		subdirPath: subdirPath,
		codec:      s.codec,
		sessionID:  "", // FIXME: fill it
		wg:         &s.wg,
	}
	return loader, nil
}

// makeSubdirPath builds a path for a filesystem direcory with instance data
func (d *defaultStorage) makeSubdirPath(desc *schema.ServiceDescription) string {
	return filepath.Join(d.cfg.DataDir, desc.GetType(), desc.GetInstance())
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

func (s *defaultStorage) ServiceMeta() ([]*storage.ServiceMeta, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
		defer s.wg.Done()
	}

	// traverse root directory and extract second level subdirectories
	// in order to obtain list of services
	subdirs1, err := ioutil.ReadDir(s.cfg.DataDir)
	if err != nil {
		return nil, err
	}
	results := make([]*storage.ServiceMeta, 0)

	for _, s1 := range subdirs1 {
		if s1.IsDir() {
			subdirs2, err := ioutil.ReadDir(s1.Name())
			if err != nil {
				return nil, err
			}
			for _, s2 := range subdirs2 {
				results = append(
					results,
					&storage.ServiceMeta{
						Description: &schema.ServiceDescription{
							Type:     s1.Name(),
							Instance: s2.Name(),
						},
					},
				)
			}
		}
	}

	return results, nil
}

func (d *defaultStorage) Quit() {
	d.cancel()
	d.wg.Wait()
}

func NewStorage(cfg *config.FilesystemStorageConfig) (storage.Storage, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &defaultStorage{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}
	return s, nil
}
