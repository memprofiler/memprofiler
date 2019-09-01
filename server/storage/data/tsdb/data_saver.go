package tsdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/data/tsdb/prometheus"
	"github.com/memprofiler/memprofiler/server/storage/metadata"

	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/tsdb/labels"
)

var _ data.Saver = (*defaultDataSaver)(nil)

// defaultDataSaver puts records to a prometheus
type defaultDataSaver struct {
	tsdbStorage     prometheus.TSDB
	metadataStorage metadata.Storage
	codec           codec
	sessionDesc     *schema.SessionDescription
	wg              *sync.WaitGroup
}

// Save store Measurement to TSDB
func (s *defaultDataSaver) Save(mm *schema.Measurement) error {
	var (
		sessionLabel = labels.Label{Name: sessionLabelName, Value: fmt.Sprintf("%v", s.SessionDescription())}
		location     = mm.GetLocations()
	)

	time, err := ptypes.Timestamp(mm.GetObservedAt())
	if err != nil {
		return err
	}

	for _, l := range location {
		appender := s.tsdbStorage.Appender()
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
	// FIXME: need some kind of atomic stop
	if err := s.metadataStorage.StopSession(context.Background(), s.sessionDesc); err != nil {
		return errors.Wrap(err, "stop session")
	}
	if err := s.tsdbStorage.Close(); err != nil {
		return errors.Wrap(err, "close TSDB storage")
	}
	return nil
}

// Session gets session identifier
func (s *defaultDataSaver) SessionDescription() *schema.SessionDescription { return s.sessionDesc }

func newDataSaver(
	sessionDesc *schema.SessionDescription,
	codec codec,
	wg *sync.WaitGroup,
	tsdbStorage prometheus.TSDB,
	metadataStorage metadata.Storage,
) (data.Saver, error) {
	saver := &defaultDataSaver{
		tsdbStorage:     tsdbStorage,
		codec:           codec,
		sessionDesc:     sessionDesc,
		metadataStorage: metadataStorage,
		wg:              wg,
	}

	return saver, nil
}
