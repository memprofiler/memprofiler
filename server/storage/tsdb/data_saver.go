package tsdb

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/tsdb/prometheus"
)

var _ storage.DataSaver = (*defaultDataSaver)(nil)

// defaultDataSaver puts records to a prometheus
type defaultDataSaver struct {
	storage     prometheus.TSDB
	codec       codec
	sessionDesc *schema.SessionDescription
	wg          *sync.WaitGroup
}

// Save store Measurement to TSDB
func (s *defaultDataSaver) Save(mm *schema.Measurement) error {
	var (
		sessionLabel = labels.Label{Name: sessionLabelName, Value: fmt.Sprintf("%v", s.SessionID())}
		location     = mm.GetLocations()
	)

	time, err := ptypes.Timestamp(mm.GetObservedAt())
	if err != nil {
		return err
	}

	for _, l := range location {
		appender := s.storage.Appender()
		mu := l.GetMemoryUsage()
		callStack, err := s.codec.encode(l.GetCallstack())
		if err != nil {
			return err
		}

		metaLabel := labels.Label{Name: metaLabelName, Value: callStack}
		measurementsInfo := MeasurementsInfo{
			{labels.Labels{sessionLabel, metaLabel, allocBytesLabel()}, float64(mu.GetAllocBytes())},
			{labels.Labels{sessionLabel, metaLabel, allocObjectsLabel()}, float64(mu.GetAllocObjects())},
			{labels.Labels{sessionLabel, metaLabel, freeBytesLabel()}, float64(mu.GetFreeBytes())},
			{labels.Labels{sessionLabel, metaLabel, freeObjectsLabel()}, float64(mu.GetFreeObjects())},
		}
		for _, mi := range measurementsInfo {
			_, err = appender.Add(mi.Labels, time.Unix(), mi.Value)
			if err != nil {
				return err
			}
		}
		err = appender.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close close data saver
func (s *defaultDataSaver) Close() error {
	defer s.wg.Done()
	return s.storage.Close()
}

// SessionID gets session identifier
func (s *defaultDataSaver) SessionID() storage.SessionID {
	return s.sessionDesc.GetSessionId()
}

func newDataSaver(
	sessionDesc *schema.SessionDescription,
	codec codec,
	wg *sync.WaitGroup,
	stor prometheus.TSDB,
) (storage.DataSaver, error) {
	saver := &defaultDataSaver{
		storage:     stor,
		codec:       codec,
		sessionDesc: sessionDesc,
		wg:          wg,
	}

	return saver, nil
}
