package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ storage.Service = (*defaultStorage)(nil)

type defaultStorage struct {
	codec codec
	cfg   *config.FilesystemStorageConfig
}

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

func (s *defaultStorage) SaveMeasurement(desc *schema.ServiceDescription, measurement *schema.Measurement) error {

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

func (d *defaultStorage) LoadAllMeasurements(
	ctx context.Context,
	desc *schema.ServiceDescription,
) (<-chan *storage.LoadResult, error) {

	// prepare list of files with measurement dumps
	subdirPath := d.makeSubdirPath(desc)
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
			case results <- d.loadSingleFile(file.Name()):
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

func (d *defaultStorage) loadSingleFile(filePath string) (result *storage.LoadResult) {
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

func (d *defaultStorage) LoadServices() ([]*schema.ServiceDescription, error) {
	panic("not implemented")
}

func (d *defaultStorage) Quit() {
	// TODO: implement graceful stop here to prevent data corruption
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

func NewStorage(cfg *config.FilesystemStorageConfig) (storage.Service, error) {
	s := &defaultStorage{
		cfg: cfg,
	}
	return s, nil
}
