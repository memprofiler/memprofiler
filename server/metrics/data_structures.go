package metrics

import (
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
	"gonum.org/v1/gonum/stat"
)

// locationStats contains raw data about memory allocations
// that happened at particular location in application code
type locationStats struct {
	inUseBytes   []float64
	inUseObjects []float64
	freeBytes    []float64
	freeObjects  []float64
	allocBytes   []float64
	allocObjects []float64
	tstamps      []*time.Time
	callStack    *schema.CallStack
}

// indicator represents the type of measurements
// For more explanation see https://golang.org/pkg/runtime/#MemProfileRecord
type indicator int8

const (
	allocObjects indicator = iota + 1
	allocBytes
	freeObjects
	freeBytes
	inUseObjects
	inUseBytes
)

var indicators = []indicator{
	allocObjects,
	allocBytes,
	freeObjects,
	freeBytes,
	inUseObjects,
	inUseBytes,
}

type computationResult struct {
	value     float64
	indicator indicator
}

// updateSeries grows times series with provided measurement
func (ls *locationStats) updateSeries(mu *schema.MemoryUsage, tstamp *time.Time) {
	ls.allocObjects = append(ls.allocObjects, float64(mu.AllocObjects))
	ls.allocBytes = append(ls.allocBytes, float64(mu.AllocBytes))
	ls.freeObjects = append(ls.freeObjects, float64(mu.FreeObjects))
	ls.freeBytes = append(ls.freeBytes, float64(mu.FreeBytes))
	ls.tstamps = append(ls.tstamps, tstamp)
}

func (ls *locationStats) toLocationMetrics(recentBorder time.Duration) *LocationMetrics {

	// 1. convert *time.Time to floats
	tstamps, recentIx := ls.makeTstamps(recentBorder)

	// 2. compute in use bytes/objects
	wg := sync.WaitGroup{}
	wg.Add(2)
	go ls.computeInUse(inUseBytes, &wg)
	go ls.computeInUse(inUseObjects, &wg)
	wg.Wait()

	// 3. compute trends of memory usage
	result := &LocationMetrics{CallStack: ls.callStack}
	result.Average = ls.computeHeapConsumptionRate(tstamps, 0)
	if recentIx != 0 {
		result.Recent = ls.computeHeapConsumptionRate(tstamps, recentIx)
	} else {
		result.Recent = result.Average
	}

	return result
}

// makeTstamps converts slice of *time.Time to float64 (convinient for regression computation)
func (ls *locationStats) makeTstamps(recentBorder time.Duration) ([]float64, int) {

	// get index of the first timestamp after (time.Now() - threshold)
	threshold := time.Now().Add(time.Duration(-1) * recentBorder)
	ix := 0

	floatTstamps := make([]float64, len(ls.tstamps))
	for i, tstamp := range ls.tstamps {
		if ix == 0 && tstamp.After(threshold) {
			ix = i
		}
		floatTstamps[i] = utils.TimeToFloat64(ls.tstamps[i])
	}

	return floatTstamps, ix
}

// computeInUse computes InUse = (Alloc - Free)
func (ls *locationStats) computeInUse(
	in indicator,
	wg *sync.WaitGroup,
) {

	defer wg.Done()
	var alloc, free []float64

	switch in {
	case inUseBytes:
		alloc, free = ls.allocBytes, ls.freeBytes
	case inUseObjects:
		alloc, free = ls.allocObjects, ls.freeObjects
	}

	inUse := make([]float64, len(alloc))
	for i := 0; i < len(inUse); i++ {
		inUse[i] = alloc[i] - free[i]
	}

	switch in {
	case inUseBytes:
		ls.inUseBytes = inUse
	case inUseObjects:
		ls.inUseObjects = inUse
	}
}

// computeHeapConsumptionRate estimates rate values for every indicator
func (ls *locationStats) computeHeapConsumptionRate(tstamps []float64, ix int) *HeapConsumptionRates {

	var (
		rates      HeapConsumptionRates
		resultChan = make(chan *computationResult, len(indicators))
	)
	for _, in := range indicators {
		go ls.computeRate(in, tstamps, ix, resultChan)
	}
	for i := 0; i < len(indicators); i++ {
		result := <-resultChan
		switch result.indicator {
		case inUseBytes:
			rates.InUseBytesRate = result.value
		case inUseObjects:
			rates.InUseObjectsRate = result.value
		case freeBytes:
			rates.FreeBytesRate = result.value
		case freeObjects:
			rates.FreeObjectsRate = result.value
		case allocBytes:
			rates.AllocBytesRate = result.value
		case allocObjects:
			rates.AllocObjectsRate = result.value
		}
	}
	return &rates
}

// computeRate performs CPU-intensive computation with time series
func (ls *locationStats) computeRate(
	in indicator,
	tstamps []float64,
	ix int,
	resultChan chan<- *computationResult,
) {
	var data []float64
	switch in {
	case inUseObjects:
		data = ls.inUseObjects
	case inUseBytes:
		data = ls.inUseBytes
	case freeObjects:
		data = ls.freeObjects
	case freeBytes:
		data = ls.freeBytes
	case allocObjects:
		data = ls.allocObjects
	case allocBytes:
		data = ls.allocBytes
	}

	_, slope := stat.LinearRegression(tstamps[ix:], data[ix:], nil, false)
	resultChan <- &computationResult{value: slope, indicator: in}
}

type sessionStats struct {
	// key - location id (hashsum from memory allocation stack)
	// value - collection of time series with heap stats
	locations map[string]*locationStats
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
		ls, exists := ss.locations[stackID]
		if !exists {
			ls = &locationStats{callStack: location.GetCallStack()}
			ss.locations[stackID] = ls
		}

		// put measurement to the time series
		ls.updateSeries(location.GetMemoryUsage(), &tstamp)
	}

	return nil
}

// FIXME: move to config
const recent = time.Second

func (ss *sessionStats) computeStatistics() []*LocationMetrics {
	results := make([]*LocationMetrics, 0, len(ss.locations))
	for _, ls := range ss.locations {
		lm := ls.toLocationMetrics(recent)
		results = append(results, lm)
	}
	return results
}

func newSessionStats() *sessionStats {
	return &sessionStats{
		locations: map[string]*locationStats{},
	}
}
