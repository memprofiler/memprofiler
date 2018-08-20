package metrics

import (
	"github.com/deckarep/golang-set"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
)

type sessionData struct {
	tstamps   []float64                // seconds since epoch
	locations map[string]*locationData // per-location stats (stackID <-> locationData)
	cfg       *config.MetricsConfig
}

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
			sdl = &locationData{window: sd.cfg.Window}
			sd.locations[l.CallStack.ID] = sdl
		}
		sdl.registerMeasurement(l.MemoryUsage)
	}

	// there may be some locations registered within a session,
	// but not within a current measurement; this means that memory
	// were allocated here before, but it has been already freed,
	// so it's necessary to put zeroes for this location
	for _, stackID := range sessionLocations.Difference(mmLocations).ToSlice() {
		sdl := sd.locations[stackID.(string)]
		sdl.registerMeasurement(emptyMemoryUsage)
	}
}

var emptyMemoryUsage = &schema.MemoryUsage{}
