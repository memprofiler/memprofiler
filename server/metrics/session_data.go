package metrics

import (
	"runtime"

	"github.com/deckarep/golang-set"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
)

type sessionData struct {
	tstamps   []float64                // seconds since epoch
	locations map[string]*locationData // per-location stats (stackID <-> locationData)
	cfg       *config.MetricsConfig
}

// registerMeasurement appends new measurement data to internal time series
func (sd *sessionData) registerMeasurement(mm *schema.Measurement) {
	// build set of stackIDs registered so far
	sessionLocations := mapset.NewSet()
	for k := range sd.locations {
		sessionLocations.Add(k)
	}

	// build set of stackIDs that came with a current message
	mmLocations := mapset.NewSet()
	for _, l := range mm.Locations {
		mmLocations.Add(l.CallStack.ID)
	}

	// iterate through incoming message and register measurements
	for _, l := range mm.Locations {
		sdl, exists := sd.locations[l.CallStack.ID]
		if !exists {
			sdl = newLocationData(l.CallStack, sd.cfg.Window)
			sd.locations[l.CallStack.ID] = sdl
		}
		sdl.registerMeasurement(l.MemoryUsage)
	}

	// there may be some locations registered within a session,
	// but not within a current measurement; this means that memory
	// were allocated here before, but it has been already freed,
	// so it's necessary to put zeroes for this location at the current timestamp
	for _, stackID := range sessionLocations.Difference(mmLocations).ToSlice() {
		sdl := sd.locations[stackID.(string)]
		sdl.registerMeasurement(emptyMemoryUsage)
	}
}

var emptyMemoryUsage = &schema.MemoryUsage{}

// computeRates performs rate computation for all known locations
func (sd *sessionData) computeRates() *schema.SessionMetrics {
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
				responseChan <- ld.computeMetrics(sd.tstamps)
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
func newSessionData(cfg *config.MetricsConfig) *sessionData {
	return &sessionData{
		locations: make(map[string]*locationData),
		cfg:       cfg,
	}
}
