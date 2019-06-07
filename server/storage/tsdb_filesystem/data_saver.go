package filesystem

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

// defaultDataSaver puts records to a file sequentially
type defaultDataSaver struct {
	tsdb        localStorage
	codec       codec
	sessionDesc *schema.SessionDescription
	cfg         *config.FilesystemStorageConfig
	wg          *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {
	sessionLabel := labels.Label{Name: SessionLabelName, Value: fmt.Sprintf("%v", s.SessionID())}

	t := mm.GetObservedAt().GetSeconds()
	v := mm.GetLocations()
	for _, v := range v {
		mu := v.GetMemoryUsage()
		ss, err := s.codec.encode(v.GetCallstack())
		if err != nil {
			return err
		}
		metaLabel := labels.Label{Name: MetaLabelName, Value: ss}

		allocBytesLabels := labels.Labels{sessionLabel, metaLabel, AllocBytesLabel}
		_, err = s.tsdb.appender().Add(allocBytesLabels, t, float64(mu.GetAllocBytes()))
		if err != nil {
			return err
		}
		allocObjectsLabels := labels.Labels{sessionLabel, metaLabel, AllocObjectsLabel}
		_, err = s.tsdb.appender().Add(allocObjectsLabels, t, float64(mu.GetAllocObjects()))
		if err != nil {
			return err
		}
		freeBytesLabels := labels.Labels{sessionLabel, metaLabel, FreeBytesLabel}
		_, err = s.tsdb.appender().Add(freeBytesLabels, t, float64(mu.GetFreeBytes()))
		if err != nil {
			return err
		}
		freeObjectsLabels := labels.Labels{sessionLabel, metaLabel, FreeObjectsLabel}
		_, err = s.tsdb.appender().Add(freeObjectsLabels, t, float64(mu.GetFreeObjects()))
		if err != nil {
			return err
		}
	}
	err := s.tsdb.appender().Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *defaultDataSaver) Close() error {
	defer s.wg.Done()
	return s.tsdb.close()
}

func (s *defaultDataSaver) SessionID() storage.SessionID { return s.sessionDesc.GetSessionId() }

func newDataSaver(
	subdirPath string,
	sessionDesc *schema.SessionDescription,
	cfg *config.FilesystemStorageConfig,
	wg *sync.WaitGroup,
	codec codec,
) (storage.DataSaver, error) {
	// create logger
	writer := log.NewSyncWriter(os.Stdout)
	logger := log.NewLogfmtLogger(writer)

	// create db
	db, err := Open(subdirPath, logger, nil, &Options{
		MinBlockDuration: model.Duration(24 * time.Hour),
		MaxBlockDuration: model.Duration(24 * time.Hour),
	})
	if err != nil {
		return nil, err
	}

	saver := &defaultDataSaver{
		tsdb:        newLocalStorage(db),
		codec:       codec,
		sessionDesc: sessionDesc,
		cfg:         cfg,
		wg:          wg,
	}

	return saver, nil
}
