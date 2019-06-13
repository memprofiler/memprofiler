package tsdb

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/tsdb/prometheus_tsdb"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

// defaultDataSaver puts records to a prometheus_tsdb
type defaultDataSaver struct {
	storage     prometheus_tsdb.Storage
	codec       codec
	sessionDesc *schema.SessionDescription
	wg          *sync.WaitGroup
}

func (s *defaultDataSaver) Save(mm *schema.Measurement) error {
	var (
		sessionLabel = labels.Label{Name: SessionLabelName, Value: fmt.Sprintf("%v", s.SessionID())}

		location = mm.GetLocations()

		wg      sync.WaitGroup
		errChan = make(chan error, len(location))
	)

	time, err := ptypes.Timestamp(mm.GetObservedAt())
	if err != nil {
		return err
	}

	for _, l := range location {
		wg.Add(1)
		go func() {
			defer wg.Done()
			appender := s.storage.Appender()
			mu := l.GetMemoryUsage()
			callStack, err := s.codec.encode(l.GetCallstack())
			if err != nil {
				errChan <- err
				return
			}

			metaLabel := labels.Label{Name: MetaLabelName, Value: callStack}
			measurementsInfo := MeasurementsInfo{
				{labels.Labels{sessionLabel, metaLabel, AllocBytesLabel}, float64(mu.GetAllocBytes())},
				{labels.Labels{sessionLabel, metaLabel, AllocObjectsLabel}, float64(mu.GetAllocObjects())},
				{labels.Labels{sessionLabel, metaLabel, FreeBytesLabel}, float64(mu.GetFreeBytes())},
				{labels.Labels{sessionLabel, metaLabel, FreeObjectsLabel}, float64(mu.GetFreeObjects())},
			}
			for _, mi := range measurementsInfo {
				_, err = appender.Add(mi.Labels, time.Unix(), mi.Value)
				if err != nil {
					errChan <- err
					return
				}
			}
			err = appender.Commit()
			if err != nil {
				errChan <- err
				return
			}
		}()
	}
	wg.Wait()

	if len(errChan) != 0 {
		var errStrings []string
		for err := range errChan {
			errStrings = append(errStrings, err.Error())
		}
		return fmt.Errorf(strings.Join(errStrings, "\n"))
	}

	return nil
}

func (s *defaultDataSaver) Close() error {
	defer s.wg.Done()
	return s.storage.Close()
}

func (s *defaultDataSaver) SessionID() storage.SessionID { return s.sessionDesc.GetSessionId() }

func newDataSaver(
	subDirPath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	wg *sync.WaitGroup,
) (storage.DataSaver, error) {
	var (
		writer = log.NewSyncWriter(os.Stdout)
		logger = log.NewLogfmtLogger(writer)
	)

	// create storage
	stor, err := prometheus_tsdb.OpenStorage(subDirPath, logger)
	if err != nil {
		return nil, err
	}

	saver := &defaultDataSaver{
		storage:     stor,
		codec:       codec,
		sessionDesc: sessionDesc,
		wg:          wg,
	}

	return saver, nil
}
