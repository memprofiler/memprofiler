package metrics

import (
	"runtime"
	"time"

	"sync"

	"context"

	"github.com/deckarep/golang-set"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
)

// sessionData contains the most recent data of the particular session;
// it's responsible for session metrics computation
type sessionData struct {
	mutex            sync.Mutex               // synchronizes access to internal structs
	locations        map[string]*locationData // per-location stats (stackID <-> locationData)
	lifetime         time.Duration            // the retention period for time series data
	sessionMetrics   *schema.SessionMetrics   // latest available session metrics (potentially outdated)
	averagingWindows []time.Duration
	outdated         bool // if metrics should be recomputed by demand
	logger           logrus.FieldLogger
}

func (sd *sessionData) populate(
	ctx context.Context,
	loadChan <-chan *storage.LoadResult,
) error {

	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	// populate session data with historical measurements coming from loader
LOOP:
	for {
		select {
		case result, ok := <-loadChan:
			if !ok {
				break LOOP
			}
			if result.Err != nil {
				sd.logger.WithError(result.Err).Error("failed to get result from loader")
			} else {
				if err := sd.appendMeasurement(result.Measurement); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// registerMeasurement appends new measurement data to internal time series
func (sd *sessionData) registerMeasurement(mm *schema.Measurement) error {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()
	return sd.appendMeasurement(mm)
}

// appendMeasurement appends new measurement data to internal time series
func (sd *sessionData) appendMeasurement(mm *schema.Measurement) error {

	// register timestamp
	timestamp, err := ptypes.Timestamp(mm.ObservedAt)
	if err != nil {
		return err
	}

	// build set of stackIDs registered so far
	sessionLocations := mapset.NewSet()
	for k := range sd.locations {
		sessionLocations.Add(k)
	}

	// build set of stackIDs that came with a current message
	mmLocations := mapset.NewSet()
	for _, l := range mm.Locations {
		mmLocations.Add(l.Callstack.Id)
	}

	// iterate through incoming message and register measurements
	for _, l := range mm.Locations {
		sdl, exists := sd.locations[l.Callstack.Id]
		if !exists {
			sdl = newLocationData(l.Callstack, sd.lifetime)
			sd.locations[l.Callstack.Id] = sdl
		}
		sdl.registerMeasurement(timestamp, l.MemoryUsage)
	}

	// there may be some locations registered within a session,
	// but not within a current measurement; this means that memory
	// was allocated in this location sometimes before, but it has been already freed,
	// so it's necessary to put zeroes for this location at the current timestamp
	for _, stackID := range sessionLocations.Difference(mmLocations).ToSlice() {
		sdl := sd.locations[stackID.(string)]
		sdl.registerMeasurement(timestamp, emptyMemoryUsage)
	}

	// mark existing sessionMetrics as outdated
	sd.outdated = true
	return nil
}

var emptyMemoryUsage = &schema.MemoryUsage{}

// getSessionMetrics returns sessionMetrics in a lazy manner
func (sd *sessionData) getSessionMetrics() *schema.SessionMetrics {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	// return existing sessionMetrics if it's not outdated
	if !sd.outdated && sd.sessionMetrics != nil {
		return sd.sessionMetrics
	}

	// otherwise perform computation from scratch
	sd.sessionMetrics = sd.computeSessionMetrics()
	sd.outdated = false
	return sd.sessionMetrics
}

// computeSessionMetrics performs rate computation for all known locations
func (sd *sessionData) computeSessionMetrics() *schema.SessionMetrics {
	var (
		requestChan  = make(chan *locationData, runtime.NumCPU())
		responseChan = make(chan *schema.LocationMetrics, runtime.NumCPU())
	)

	// rate computation is a CPU-bound operation, so spread it across cores
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				ld, ok := <-requestChan
				if !ok {
					return
				}
				responseChan <- ld.computeMetrics(sd.averagingWindows)
			}
		}()
	}

	// enqueue jobs
	go func() {
		for _, ld := range sd.locations {
			requestChan <- ld
		}
		close(requestChan) // notify consumers
	}()

	// await results
	results := make([]*schema.LocationMetrics, len(sd.locations))
	for i := 0; i < len(sd.locations); i++ {
		results[i] = <-responseChan
	}

	return &schema.SessionMetrics{Locations: results}
}

// newSessionData instantiates new
func newSessionData(logger logrus.FieldLogger, averagingWindows []time.Duration) *sessionData {
	return &sessionData{
		locations:        make(map[string]*locationData),
		lifetime:         averagingWindows[len(averagingWindows)-1],
		averagingWindows: averagingWindows,
		logger:           logger,
	}
}
