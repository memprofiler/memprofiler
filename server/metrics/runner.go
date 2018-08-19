package metrics

import (
	"context"
	"sync"

	"sort"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ Runner = (*defaultRunner)(nil)

type defaultRunner struct {
	mutex    sync.RWMutex
	sessions map[string]*sessionStats
	logger   logrus.FieldLogger
	storage  storage.Storage
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func (r *defaultRunner) PutMeasurement(sd *storage.SessionDescription, mm *schema.Measurement) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ss, exists := r.sessions[sd.String()]
	if !exists {
		ss = newSessionStats()
		r.sessions[sd.String()] = ss
	}

	ss.registerMeasurement(mm)
	return nil
}

func (r *defaultRunner) GetSessionMetrics(
	ctx context.Context,
	sd *storage.SessionDescription,
) (*schema.SessionMetrics, error) {

	r.mutex.RLock()
	ss, exists := r.sessions[sd.String()]
	r.mutex.RUnlock()
	if exists {
		result := &schema.SessionMetrics{
			Locations: ss.computeStatistics(),
		}
		return result, nil
	}

	return r.generateSessionMetrics(ctx, sd)
}

func (r *defaultRunner) generateSessionMetrics(
	ctx context.Context,
	sd *storage.SessionDescription,
) (*schema.SessionMetrics, error) {

	// prepare
	dataLoader, err := r.storage.NewDataLoader(sd)
	if err != nil {
		return nil, err
	}

	r.wg.Add(1)
	defer func() {
		if err := dataLoader.Close(); err != nil {
			r.logger.WithError(err).Error("Failed to close data loader")
		}
		r.wg.Done()
	}()

	ss := newSessionStats()
	loadChan, err := dataLoader.Load(ctx)
	if err != nil {
		return nil, err
	}

	// populate stats with data coming from loader
LOOP:
	for {
		select {
		case data, ok := <-loadChan:
			if !ok {
				break LOOP
			}
			if data.Err != nil {
				r.logger.WithError(data.Err).Error("failed to get data from loader")
			} else {
				ss.registerMeasurement(data.Measurement)
			}
		case <-r.ctx.Done():
			return nil, r.ctx.Err()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// by default sort by InUseBytes, because this tends to be the most import indicator
	locations := ss.computeStatistics()
	sort.Slice(locations, func(i, j int) bool {
		// descending order
		return locations[i].Average.InUseBytesRate > locations[j].Average.InUseBytesRate
	})

	result := &schema.SessionMetrics{Locations: locations}
	return result, nil
}

func (r *defaultRunner) Quit() {
	r.cancel()
	r.wg.Wait()
}

// NewRunner instantiates new runner
func NewRunner(logger logrus.FieldLogger) Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &defaultRunner{
		logger: logger,
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}
