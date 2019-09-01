package filesystem

import (
	"context"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/metadata"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
)

var _ data.Saver = (*defaultDataSaver)(nil)

// defaultDataSaver puts records to a file sequentially
type defaultDataSaver struct {
	codec           codec
	delimiter       []byte
	fd              *os.File // data file
	metadataStorage metadata.Storage
	sessionDesc     *schema.SessionDescription
	cfg             *config.FilesystemStorageConfig
	wg              *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {

	// put delimiter after last record
	if _, err := s.fd.Write(s.delimiter); err != nil {
		return err
	}

	// serialize measurement into the file
	if err := s.codec.encode(s.fd, mm); err != nil {
		return err
	}

	// sync file if required
	if s.cfg.SyncWrite {
		if err := s.fd.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (s *defaultDataSaver) Close() error {
	defer func() {
		s.wg.Done()
	}()
	if err := s.metadataStorage.StopSession(context.Background(), s.sessionDesc); err != nil {
		return errors.Wrap(err, "stop sessionDesc")
	}
	if err := s.fd.Close(); err != nil {
		return errors.Wrap(err, "close file descriptor")
	}
	return nil
}

func (s *defaultDataSaver) SessionDescription() *schema.SessionDescription { return s.sessionDesc }

func newDataSaver(
	dataFilePath string,
	sessionDesc *schema.SessionDescription,
	cfg *config.FilesystemStorageConfig,
	wg *sync.WaitGroup,
	codec codec,
	metadataStorage metadata.Storage,
) (data.Saver, error) {

	// open file to store records
	fd, err := os.OpenFile(dataFilePath, os.O_WRONLY|os.O_CREATE, filePermissions)
	if err != nil {
		return nil, err
	}

	saver := &defaultDataSaver{
		fd:              fd,
		delimiter:       []byte{10}, // '\n'
		codec:           codec,
		sessionDesc:     sessionDesc,
		cfg:             cfg,
		wg:              wg,
		metadataStorage: metadataStorage,
	}

	return saver, nil
}
