package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"

	"path/filepath"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/prometheus/tsdb/labels"
	"github.com/sirupsen/logrus"
)

type defaultDataLoader struct {
	tsdb   localStorage
	codec  codec
	sd     *schema.SessionDescription
	logger logrus.FieldLogger
	wg     *sync.WaitGroup
}

const (
	loadChanCapacity = 256
)

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {
	// prepare bufferized channel for results
	results := make(chan *storage.LoadResult, loadChanCapacity)

	querier, err := l.tsdb.querier(context.Background(), 20, 30)
	if err != nil {
		return nil, err
	}

	allocObjectsSeriesSet, _ := querier.Select([]labels.Matcher{
		labels.NewEqualMatcher(SessionLabelName, fmt.Sprintf("%v", l.sd.GetSessionId())),
		labels.NewEqualMatcher(MetricTypeLabelName, fmt.Sprintf("%v", "AllocObjects")),
	}...)

	// scan records line by line
	go func() {
		defer close(results)
		for allocObjectsSeriesSet.Next() {
			series := allocObjectsSeriesSet.At()
			labs := series.Labels()
			var receiver schema.Callstack
			err := l.codec.decode(labs.Get(MetaLabelName), &receiver)
			seriesIterator := series.Iterator()
			for seriesIterator.Next() {
				writeTime, val := seriesIterator.At()
				rr := &storage.LoadResult{
					Measurement: &schema.Measurement{
						Locations: []*schema.Location{
							{
								MemoryUsage: &schema.MemoryUsage{
									AllocObjects: int64(val),
								},
								Callstack: &receiver,
							},
						},
						ObservedAt: writeTime,
					},
					Err: err,
				}
				select {
				case results <- rr:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return results, nil
}

// loadMeasurement disk
func (l *defaultDataLoader) loadMeasurement(data []byte) *storage.LoadResult {
	var receiver schema.Measurement
	err := l.codec.decode(bytes.NewReader(data), &receiver)
	return &storage.LoadResult{Measurement: &receiver, Err: err}
}

func (l *defaultDataLoader) Close() error {
	defer l.wg.Done()
	return l.fd.Close()
}

func newDataLoader(
	subdirPath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger logrus.FieldLogger,
	wg *sync.WaitGroup,
) (storage.DataLoader, error) {

	// open file to load records
	filename := filepath.Join(subdirPath, "data")
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	contextLogger := logger.WithFields(logrus.Fields{
		"type":        sessionDesc.GetServiceType(),
		"instance":    sessionDesc.GetServiceInstance(),
		"sessionDesc": storage.SessionIDToString(sessionDesc.GetSessionId()),
		"measurement": filename,
	})

	loader := &defaultDataLoader{
		sd:     sessionDesc,
		fd:     fd,
		codec:  codec,
		logger: contextLogger,
		wg:     wg,
	}
	return loader, nil
}
