package metrics

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ Computer = (*defaultComputer)(nil)

type defaultComputer struct {
	logger *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ComputeSessionMetrics caches whole session data and computes statistics;
// perhaps it worth make same effort to limit memory consumption
func (c *defaultComputer) ComputeSessionMetrics(
	ctx context.Context,
	dataLoader storage.DataLoader,
) ([]*LocationMetrics, error) {

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
				c.logger.WithError(err).Error("Failed to get data from loader")
			} else {
				ss.registerMeasurement(data.Measurement)
			}
		case <-c.ctx.Done():
			return nil, c.ctx.Err()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, nil
}

func (c *defaultComputer) Quit() {
	c.cancel()
	c.wg.Wait()
}

// New instantiates new Computer
func New(logger *logrus.Logger) Computer {
	ctx, cancel := context.WithCancel(context.Background())
	return &defaultComputer{
		logger: logger,
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}
