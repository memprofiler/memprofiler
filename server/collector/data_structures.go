package collector

import (
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
)

type heapStats struct {
	inUseObjects []float64
	inUseBytes   []float64
	allocObjects []float64
	allocBytes   []float64
	tstamps      []float64
}

// updateSeries grows times series with provided measurement
func (hs *heapStats) updateSeries(mu *schema.MemoryUsage, tstamp time.Time) {
	hs.inUseObjects = append(hs.inUseObjects, float64(mu.InUseObjects))
	hs.inUseBytes = append(hs.inUseBytes, float64(mu.InUseBytes))
	hs.allocObjects = append(hs.allocObjects, float64(mu.AllocObjects))
	hs.allocBytes = append(hs.allocBytes, float64(mu.AllocBytes))
	hs.tstamps = append(hs.tstamps, utils.TimeToFloat64(tstamp))
}

type instanceStats struct {
	mutex *sync.RWMutex

	// key - location id (hashsum from memory allocation stack)
	// value - collection of time series with heap stats
	data map[string]*heapStats

	// key - location id (hashsum from memory allocation stack)
	// value - memory allocation stack
	locations map[string]*schema.CallStack
}

func (is *instanceStats) registerMeasurement(mm *schema.Measurement) error {

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
		data, exists := is.data[stackID]
		if !exists {
			data = &heapStats{}
			is.data[stackID] = data
		}

		// register location
		if _, exists := is.locations[stackID]; !exists {
			is.locations[stackID] = location.GetCallStack()
		}

		// put measurement to the time series
		data.updateSeries(location.GetMemoryUsage(), tstamp)
	}

	return nil
}

type serviceStats struct {
	// key - service instance id
	// values - instance stats
	instances map[string]*instanceStats
}

type overallStats struct {
	// key - service type id
	// values - service stats
	services map[string]*serviceStats
}

func newInstanceStats() *instanceStats {
	return &instanceStats{
		mutex:     &sync.RWMutex{},
		data:      map[string]*heapStats{},
		locations: map[string]*schema.CallStack{},
	}
}

func newServiceStats() *serviceStats {
	return &serviceStats{
		instances: map[string]*instanceStats{},
	}
}

func newOverallStats() *overallStats {
	return &overallStats{
		services: map[string]*serviceStats{},
	}
}
