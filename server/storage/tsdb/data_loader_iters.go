package tsdb

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
	localTSDB "github.com/memprofiler/memprofiler/server/storage/tsdb/prometheus_tsdb"
)

// MeasurementIterator gets measurement for each time in TSDB
type MeasurementIterator struct {
	querier tsdb.Querier
	codec   codec

	// current data state
	currentTime            int64
	MemoryUsageIteratorMap map[string]*MemoryUsageIterator
}

// Next check for next element
func (i *MeasurementIterator) Next() bool {
	if len(i.MemoryUsageIteratorMap) > 0 {
		i.updateMin()
		return true
	}
	return false
}

// At get current measurement from state
func (i *MeasurementIterator) At() *schema.Measurement {
	// TODO: use nsec
	t, err := ptypes.TimestampProto(time.Unix(i.currentTime, 0))
	if err != nil {
		panic(err)
	}

	return &schema.Measurement{
		ObservedAt: t,
		Locations:  i.currentLocations(),
	}
}

func (i *MeasurementIterator) updateMin() {
	// reset time to time now for get next min
	i.currentTime = time.Now().Unix()

	for _, v := range i.MemoryUsageIteratorMap {
		if v.currentTime < i.currentTime {
			i.currentTime = v.currentTime
		}
	}
}

func (i *MeasurementIterator) currentLocations() []*schema.Location {
	var currentLocations []*schema.Location

	// create Locations
	for l, v := range i.MemoryUsageIteratorMap {
		if v.currentTime == i.currentTime {
			currentLocations = append(currentLocations, i.getLocation(l, v.CurrentMemoryUsage()))

			// delete MemoryUsageIterator if no new values
			if !i.MemoryUsageIteratorMap[l].Next() {
				delete(i.MemoryUsageIteratorMap, l)
			}
		}
	}

	return currentLocations
}

func (i *MeasurementIterator) getLocation(callStack string, memUsage *schema.MemoryUsage) *schema.Location {
	cs := &schema.Callstack{}

	err := i.codec.decode(callStack, cs)
	if err != nil {
		panic(err)
	}

	return &schema.Location{
		MemoryUsage: memUsage,
		Callstack:   cs,
	}
}

// NewMeasurementIterator iterator over measurements in session
func NewMeasurementIterator(tsdb localTSDB.TSDB, codec codec, sessionLabel labels.Label) (*MeasurementIterator, error) {
	querier, err := tsdb.Querier(0, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	// TODO: use LabelValuesFor when implemented
	// metaLabels, err := querier.LabelValuesFor(MetaLabelName, sessionLabel)
	metaLabels, err := querier.LabelValues(MetaLabelName)
	if err != nil {
		return nil, err
	}

	locationsIterMap := make(map[string]*MemoryUsageIterator, len(metaLabels))

	for _, m := range metaLabels {
		metaLabel := labels.Label{Name: MetaLabelName, Value: m}
		mui, ok := NewMemoryUsageIterator(querier, sessionLabel, metaLabel)
		if ok {
			locationsIterMap[m] = mui
		}
	}
	li := &MeasurementIterator{
		querier:                querier,
		currentTime:            time.Now().Unix(),
		MemoryUsageIteratorMap: locationsIterMap,
		codec:                  codec,
	}
	return li, nil
}

// MemoryUsageIterator memory usage statistic iterator for single location
type MemoryUsageIterator struct {
	// state of iteration
	currentTime     int64
	currentMemUsage *schema.MemoryUsage

	// iterator for each memory usage type
	allocObjectsIterator tsdb.SeriesIterator
	allocBytesIterator   tsdb.SeriesIterator
	freeObjectsIterator  tsdb.SeriesIterator
	freeBytesIterator    tsdb.SeriesIterator
}

// CurrentTime current time for location
func (i *MemoryUsageIterator) CurrentTime() int64 {
	return i.currentTime
}

// CurrentMemoryUsage current memory usage for location
func (i *MemoryUsageIterator) CurrentMemoryUsage() *schema.MemoryUsage {
	return i.currentMemUsage
}

// Next check for next element and update state
func (i *MemoryUsageIterator) Next() bool {
	next := i.allocObjectsIterator.Next() &&
		i.allocBytesIterator.Next() &&
		i.freeObjectsIterator.Next() &&
		i.freeBytesIterator.Next()

	if next {
		t1, allocObjects := i.allocObjectsIterator.At()
		t2, allocBytes := i.allocBytesIterator.At()
		t3, freeObjects := i.freeObjectsIterator.At()
		t4, freeBytes := i.freeBytesIterator.At()

		if !(t1 == t2 && t2 == t3 && t3 == t4) {
			panic("time for measurement is incorrect")
		}

		// set current time
		i.currentTime = t1

		// set current memory usage
		i.currentMemUsage = &schema.MemoryUsage{
			AllocObjects: int64(allocObjects),
			AllocBytes:   int64(allocBytes),
			FreeObjects:  int64(freeObjects),
			FreeBytes:    int64(freeBytes),
		}
	}

	return next
}

// NewMemoryUsageIterator iterator for simple Location
func NewMemoryUsageIterator(querier tsdb.Querier, sessionLabel, metaLabel labels.Label) (*MemoryUsageIterator, bool) {
	allocObjectsIterator, ok := createSeriesIterator(querier, sessionLabel, metaLabel, AllocObjectsLabel)
	if !ok {
		return nil, false
	}
	allocBytesIterator, ok := createSeriesIterator(querier, sessionLabel, metaLabel, AllocBytesLabel)
	if !ok {
		return nil, false
	}
	freeObjectsIterator, ok := createSeriesIterator(querier, sessionLabel, metaLabel, FreeObjectsLabel)
	if !ok {
		return nil, false
	}
	freeBytesIterator, ok := createSeriesIterator(querier, sessionLabel, metaLabel, FreeBytesLabel)
	if !ok {
		return nil, false
	}

	mui := &MemoryUsageIterator{
		allocObjectsIterator: allocObjectsIterator,
		allocBytesIterator:   allocBytesIterator,
		freeObjectsIterator:  freeObjectsIterator,
		freeBytesIterator:    freeBytesIterator,
	}

	// initial call Next
	next := mui.Next()
	if !next {
		return nil, false
	}

	return mui, true
}

func createSeriesIterator(querier tsdb.Querier, sessionLabel, metaLabel, l labels.Label) (tsdb.SeriesIterator, bool) {
	seriesSet, _ := querier.Select([]labels.Matcher{
		labels.NewEqualMatcher(sessionLabel.Name, sessionLabel.Value),
		labels.NewEqualMatcher(metaLabel.Name, metaLabel.Value),
		labels.NewEqualMatcher(l.Name, l.Value),
	}...)

	// we iterate once because must be single series for current labels
	next := seriesSet.Next()
	if !next {
		return nil, false
	}
	series := seriesSet.At()

	return series.Iterator(), true
}
