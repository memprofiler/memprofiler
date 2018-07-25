package collector

import (
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
	"gonum.org/v1/gonum/stat"
)

type series struct {
	values []float64
	slope  float64
}

func (s *series) add(value float64) {
	// append value to array
	s.values = append(s.values, value)
}

// heapStats contains raw data and statistics for
// heap usage indicators provided by runtime
type heapStats struct {
	inUseObjects *series
	inUseBytes   *series
	allocObjects *series
	allocBytes   *series
	tstamps      *series
}

// updateSeries grows times series with provided measurement
func (hs *heapStats) updateSeries(mu *schema.MemoryUsage, tstamp time.Time) {
	hs.inUseObjects.add(float64(mu.InUseObjects))
	hs.inUseBytes.add(float64(mu.InUseBytes))
	hs.allocObjects.add(float64(mu.AllocObjects))
	hs.allocBytes.add(float64(mu.AllocBytes))
	hs.tstamps.add(utils.TimeToFloat64(tstamp))
}

func (hs *heapStats) computeRegression() {
	_, hs.inUseObjects.slope = stat.LinearRegression(hs.tstamps.values, hs.inUseObjects.values, nil, false)
	_, hs.inUseBytes.slope = stat.LinearRegression(hs.tstamps.values, hs.inUseBytes.values, nil, false)
	_, hs.allocObjects.slope = stat.LinearRegression(hs.tstamps.values, hs.allocObjects.values, nil, false)
	_, hs.allocBytes.slope = stat.LinearRegression(hs.tstamps.values, hs.allocBytes.values, nil, false)
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
		data.computeRegression()
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
