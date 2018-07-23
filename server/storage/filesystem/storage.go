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

var _ storage.Service = (*defaultStorage)(nil)

// defaultStorage uses filesystem for a file
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

func (s *defaultStorage) SaveMeasurement(desc *schema.ServiceDescription, measurement *schema.Measurement) error {

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		s.wg.Add(1)
		defer s.wg.Done()
	}

	// obtain directory to store data coming from a particular service instance
	subdirPath := s.makeSubdirPath(desc)
	if _, err := os.Stat(subdirPath); err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		// create directory if it doesn't exist
		if err = os.MkdirAll(subdirPath, dirPermissions); err != nil {
			return fmt.Errorf("failed to create directory for service data: %v", err)
		}
	}

	// open file for writing
	filePath, err := s.makeFilePath(subdirPath, measurement.GetObservedAt())
	if err != nil {
		return fmt.Errorf("failed to make path for file to store measurement: %v", err)
	}
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, filePermissions)
	if err != nil {
		return err
	}
	defer fd.Close()

	// serialize measurement into the file
	if err := s.codec.encode(fd, measurement); err != nil {
		return err
	}

	// sync file if needed
	if s.cfg.SyncWrite {
		if err := fd.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (s *defaultStorage) LoadAllMeasurements(
	ctx context.Context,
	desc *schema.ServiceDescription,
) (<-chan *storage.LoadResult, error) {

	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		s.wg.Add(1)
		defer s.wg.Done()
	}

	// prepare list of files with measurement dumps
	subdirPath := s.makeSubdirPath(desc)
	files, err := ioutil.ReadDir(subdirPath)
	if err != nil {
		return nil, err
	}

	// take files from disk, deserialize it and send it to the caller asynchronously
	results := make(chan *storage.LoadResult)
	go func() {
		defer close(results)
		for _, file := range files {
			select {
			case results <- s.loadSingleMeasurement(file.Name()):
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

// loadSingleMeasurement reads single measurement from file
func (d *defaultStorage) loadSingleMeasurement(filePath string) (result *storage.LoadResult) {
	result = &storage.LoadResult{}

	fd, err := os.OpenFile(filePath, os.O_RDONLY, filePermissions)
	if err != nil {
		result.Err = err
		return
	}
	defer fd.Close()

	var measurement schema.Measurement
	if err := d.codec.decode(fd, &measurement); err != nil {
		result.Err = err
		return
	}

	result.Measurement = &measurement
	return
}

// makeSubdirPath builds a path for a filesystem direcory with instance data
func (d *defaultStorage) makeSubdirPath(desc *schema.ServiceDescription) string {
	return filepath.Join(d.cfg.DataDir, desc.GetType(), desc.GetInstance())
}

const timeFormat = "%d%02d%02d-%02d%02d%02d"

// makeFilePath creates a path for a file to store a distinct measurement
func (d *defaultStorage) makeFilePath(subdirPath string, tstamp *timestamp.Timestamp) (string, error) {

	t, err := ptypes.Timestamp(tstamp)
	if err != nil {
		return "", err
	}

	dump := fmt.Sprintf(timeFormat,
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return filepath.Join(subdirPath, dump), nil
}

func (s *defaultStorage) LoadServices() ([]*schema.ServiceDescription, error) {

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
	results := make([]*schema.ServiceDescription, 0)

	for _, s1 := range subdirs1 {
		if s1.IsDir() {
			subdirs2, err := ioutil.ReadDir(s1.Name())
			if err != nil {
				return nil, err
			}
			for _, s2 := range subdirs2 {
				results = append(
					results,
					&schema.ServiceDescription{Type: s1.Name(), Instance: s2.Name()},
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

func NewStorage(cfg *config.FilesystemStorageConfig) (storage.Service, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &defaultStorage{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}
	return s, nil
}
