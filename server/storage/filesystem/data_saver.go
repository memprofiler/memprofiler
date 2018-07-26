package filesystem

import (
	"fmt"
	"os"
	"sync"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

type defaultDataSaver struct {
	subdirPath string
	codec      codec
	sessionID  storage.SessionID
	cfg        *config.FilesystemStorageConfig
	wg         *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {

	// open file for writing
	filePath, err := makeFilePath(s.subdirPath, mm.GetObservedAt())
	if err != nil {
		return fmt.Errorf("failed to make path for file to store measurement: %v", err)
	}
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, filePermissions)
	if err != nil {
		return err
	}
	defer fd.Close()

	// serialize measurement into the file
	if err := s.codec.encode(fd, mm); err != nil {
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

func (s *defaultDataSaver) Close() { s.wg.Done() }

func (s *defaultDataSaver) SessionID() storage.SessionID { return s.sessionID }
