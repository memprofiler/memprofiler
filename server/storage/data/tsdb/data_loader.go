package tsdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/tsdb/labels"
	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/data/tsdb/prometheus"
)

const (
	loadChanCapacity = 256
)

type defaultDataLoader struct {
	storage prometheus.TSDB
	codec   codec
	sd      *schema.SessionDescription
	logger  *zerolog.Logger
	wg      *sync.WaitGroup
}

// Load read data from TSDB
func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *data.LoadResult, error) {
	var sessionLabel = labels.Label{
		Name:  sessionLabelName,
		Value: fmt.Sprintf("%d", l.sd.GetId()),
	}

	li, err := NewMeasurementIterator(l.storage, l.codec, sessionLabel)
	if err != nil {
		return nil, err
	}

	// prepare bufferized channel for results
	results := make(chan *data.LoadResult, loadChanCapacity)
	go func() {
		defer close(results)
		for li.Next() {
			var (
				measurement = li.At()
				err         = li.Error()
			)

			m := &data.LoadResult{Measurement: measurement, Err: err}

			select {
			case results <- m:
			case <-ctx.Done():
				break
			}
		}
	}()

	return results, nil
}

func (l *defaultDataLoader) Close() error {
	defer l.wg.Done()
	return nil
}

func newDataLoader(
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger *zerolog.Logger,
	wg *sync.WaitGroup,
	stor prometheus.TSDB,
) (data.Loader, error) {
	// open file to load records
	contextLogger := logger.With().Fields(map[string]interface{}{
		"service":    sessionDesc.InstanceDescription.ServiceName,
		"instance":   sessionDesc.InstanceDescription.InstanceName,
		"session_id": sessionDesc.Id,
	}).Logger()

	loader := &defaultDataLoader{
		storage: stor,
		sd:      sessionDesc,
		codec:   codec,
		logger:  &contextLogger,
		wg:      wg,
	}
	return loader, nil
}
