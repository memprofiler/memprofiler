package tsdb

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/tsdb/labels"
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
	localTSDB "github.com/memprofiler/memprofiler/server/storage/tsdb/prometheus_tsdb"
)

const (
	loadChanCapacity = 256
)

type defaultDataLoader struct {
	storage localTSDB.TSDB
	codec   codec
	sd      *schema.SessionDescription
	logger  logrus.FieldLogger
	wg      *sync.WaitGroup
}

// Load read data from TSDB
func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {
	var sessionLabel = labels.Label{Name: SessionLabelName, Value: fmt.Sprintf("%v", l.sd.GetSessionId())}

	li, err := NewMeasurementIterator(l.storage, l.codec, sessionLabel)
	if err != nil {
		return nil, err
	}

	// prepare bufferized channel for results
	results := make(chan *storage.LoadResult, loadChanCapacity)
	go func() {
		defer close(results)
		for li.Next() {
			m := &storage.LoadResult{Measurement: li.At(), Err: err}

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
	subdirPath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger logrus.FieldLogger,
	wg *sync.WaitGroup,
) (storage.DataLoader, error) {
	// TODO: wrap to logrus interface
	var (
		writer  = log.NewSyncWriter(os.Stdout)
		logger2 = log.NewLogfmtLogger(writer)
	)

	// open file to load records
	contextLogger := logger.WithFields(logrus.Fields{
		"type":        sessionDesc.GetServiceType(),
		"instance":    sessionDesc.GetServiceInstance(),
		"sessionDesc": storage.SessionIDToString(sessionDesc.GetSessionId()),
	})

	// create storage
	stor, err := localTSDB.OpenTSDB(subdirPath, logger2)
	if err != nil {
		return nil, err
	}

	loader := &defaultDataLoader{
		storage: stor,
		sd:      sessionDesc,
		codec:   codec,
		logger:  contextLogger,
		wg:      wg,
	}
	return loader, nil
}