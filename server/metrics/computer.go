package metrics

import (
	"context"
	"sync"

	"sort"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ Computer = (*defaultComputer)(nil)

type defaultComputer struct {
	logger logrus.FieldLogger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// SessionMetrics caches whole session data and computes statistics;
// perhaps it worth make same effort to limit memory consumption
func (c *defaultComputer) SessionMetrics(
	ctx context.Context,
	dataLoader storage.DataLoader,
) (*schema.SessionMetrics, error) {

	c.wg.Add(1)
	defer func() {
		if err := dataLoader.Close(); err != nil {
			c.logger.WithError(err).Error("Failed to close data loader")
		}
		c.wg.Done()
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
				c.logger.WithError(data.Err).Error("Failed to get data from loader")
			} else {
				ss.registerMeasurement(data.Measurement)
			}
		case <-c.ctx.Done():
			return nil, c.ctx.Err()
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

func (c *defaultComputer) Quit() {
	c.cancel()
	c.wg.Wait()
}

// New instantiates new Computer
func New(logger logrus.FieldLogger) Computer {
	ctx, cancel := context.WithCancel(context.Background())
	return &defaultComputer{
		logger: logger,
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}
