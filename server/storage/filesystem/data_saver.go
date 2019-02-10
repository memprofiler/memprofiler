package filesystem

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

var delimiter = []byte{10} // '\n'

// defaultDataSaver puts records to a file sequentially
type defaultDataSaver struct {
	codec              codec
	cache              cache
	fd                 *os.File
	sessionDescription *storage.SessionDescription
	mmID               measurementID
	cfg                *config.FilesystemStorageConfig
	wg                 *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {

	// put delimiter after last record
	if _, err := s.fd.Write(delimiter); err != nil {
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

	// put record to cache
	if s.cache != nil {
		mmMeta := &measurementMetadata{
			session: s.sessionDescription,
			mmID:    s.mmID,
		}
		s.cache.put(mmMeta, mm)
	}

	// increment measurement ID
	s.mmID++

	return nil
}

func (s *defaultDataSaver) Close() error {
	defer s.wg.Done()
	return s.fd.Close()
}

func (s *defaultDataSaver) SessionID() storage.SessionID { return s.sessionDescription.SessionID }

func newDataSaver(
	subdirPath string,
	serviceDescription *schema.ServiceDescription,
	sessionID storage.SessionID,
	cfg *config.FilesystemStorageConfig,
	wg *sync.WaitGroup,
	codec codec,
	cache cache,
) (storage.DataSaver, error) {

	// open file to store records
	filename := filepath.Join(subdirPath, "data")
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, filePermissions)
	if err != nil {
		return nil, err
	}

	saver := &defaultDataSaver{
		fd:    fd,
		codec: codec,
		cache: cache,
		sessionDescription: &storage.SessionDescription{
			ServiceDescription: serviceDescription,
			SessionID:          sessionID,
		},
		cfg: cfg,
		wg:  wg,
	}

	return saver, nil
}
