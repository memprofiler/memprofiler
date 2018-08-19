package filesystem

import (
	"os"
	"sync"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

type defaultDataSaver struct {
	codec codec
	cache cache

	subdirPath         string
	serviceDescription *schema.ServiceDescription
	sessionID          storage.SessionID
	mmID               measurementID

	cfg *config.FilesystemStorageConfig
	wg  *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {

	// open file for writing
	filePath := makeFilename(s.subdirPath, s.mmID)
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, filePermissions)
	if err != nil {
		return err
	}
	defer fd.Close()

	// serialize measurement into the file
	if err := s.codec.encode(fd, mm); err != nil {
		return err
	}

	// sync file if required
	if s.cfg.SyncWrite {
		if err := fd.Sync(); err != nil {
			return err
		}
	}

	// put record to cache
	if s.cache != nil {
		mmMeta := &measurementMetadata{
			serviceDescription: s.serviceDescription,
			sessionID:          s.sessionID,
			mmID:               s.mmID,
		}
		s.cache.put(mmMeta, mm)
	}

	// increment measurement ID
	s.mmID++

	return nil
}

func (s *defaultDataSaver) Close() error {
	s.wg.Done()
	return nil
}

func (s *defaultDataSaver) SessionID() storage.SessionID { return s.sessionID }
