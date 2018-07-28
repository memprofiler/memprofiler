package metrics

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
)

// locationStats contains raw data about memory allocations
// that happend in a particular location
type locationStats struct {
	inUseObjects []float64
	inUseBytes   []float64
	allocObjects []float64
	allocBytes   []float64
	tstamps      []*time.Time
	callStack    *schema.CallStack
}

// updateSeries grows times series with provided measurement
func (hs *locationStats) updateSeries(mu *schema.MemoryUsage, tstamp *time.Time) {
	hs.inUseObjects = append(hs.inUseObjects, float64(mu.InUseObjects))
	hs.inUseBytes = append(hs.inUseBytes, float64(mu.InUseBytes))
	hs.allocObjects = append(hs.allocObjects, float64(mu.AllocObjects))
	hs.allocBytes = append(hs.allocBytes, float64(mu.AllocBytes))
	hs.tstamps = append(hs.tstamps, tstamp)
}

func (hs *locationStats) computeRegression() {
	// _, hs.inUseObjects.slope = stat.LinearRegression(hs.tstamps.values, hs.inUseObjects.values, nil, false)
	// _, hs.inUseBytes.slope = stat.LinearRegression(hs.tstamps.values, hs.inUseBytes.values, nil, false)
	// _, hs.allocObjects.slope = stat.LinearRegression(hs.tstamps.values, hs.allocObjects.values, nil, false)
	// _, hs.allocBytes.slope = stat.LinearRegression(hs.tstamps.values, hs.allocBytes.values, nil, false)
}

type sessionStats struct {
	// key - location id (hashsum from memory allocation stack)
	// value - collection of time series with heap stats
	data map[string]*locationStats
}

func (ss *sessionStats) registerMeasurement(mm *schema.Measurement) error {

	tstamp, err := ptypes.Timestamp(mm.GetObservedAt())
	if err != nil {
		return err
	}

	for _, location := range mm.GetLocations() {

		// get stack unique identifier
		stackID, err := utils.HashCallStack(location.GetCallStack())
		if err != nil {
			return err
		}

		// register data
		data, exists := ss.data[stackID]
		if !exists {
			data = &locationStats{
				callStack: location.GetCallStack(),
			}
			ss.data[stackID] = data
		}

		// put measurement to the time series
		data.updateSeries(location.GetMemoryUsage(), &tstamp)
	}
	return nil
}

func newSessionStats() *sessionStats {
	return &sessionStats{
		data: map[string]*locationStats{},
	}
}
