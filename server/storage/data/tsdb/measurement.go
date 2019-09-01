package tsdb

import (
	"time"

	"github.com/memprofiler/memprofiler/server/storage/data/tsdb/prometheus"

	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
)

// MeasurementIterator gets measurement for each time in TSDB
type MeasurementIterator interface {
	Next() bool
	At() *schema.Measurement
	Error() error
}

var _ MeasurementIterator = (*measurementIterator)(nil)

type measurementIterator struct {
	querier tsdb.Querier
	codec   codec

	// current data state
	currentTime            int64
	memoryUsageIteratorMap map[string]MemoryUsageIterator
	error                  error
}

// Next check for next element
func (i *measurementIterator) Next() bool {
	// do iteration
	// if memory usage map contains records, we must get minimum time for current iteration
	if len(i.memoryUsageIteratorMap) > 0 {
		i.updateMin()
		return true
	}
	return false
}

// At get current measurement from state
func (i *measurementIterator) At() *schema.Measurement {
	// TODO: use nano seconds instead of milliseconds
	t, err := ptypes.TimestampProto(time.Unix(i.currentTime, 0))
	if err != nil {
		i.error = err
	}

	location, err := i.currentLocations()
	if err != nil {
		i.error = err
	}

	return &schema.Measurement{
		ObservedAt: t,
		Locations:  location,
	}
}

// At get current measurement from state
func (i *measurementIterator) Error() error {
	return i.error
}

func (i *measurementIterator) updateMin() {
	// reset time to time now for get next min
	i.currentTime = time.Now().Unix()

	// get minimum time from all measurements
	for _, v := range i.memoryUsageIteratorMap {
		memoryUsageTimeState, _ := v.At()
		if memoryUsageTimeState < i.currentTime {
			i.currentTime = memoryUsageTimeState
		}
	}
}

func (i *measurementIterator) currentLocations() ([]*schema.Location, error) {
	var currentLocations []*schema.Location

	// get all Sessions from map with same time (current time state)
	for l, v := range i.memoryUsageIteratorMap {
		memoryUsageTimeState, currentMemoryUsage := v.At()
		if memoryUsageTimeState == i.currentTime {
			location, err := i.getLocation(l, currentMemoryUsage)
			if err != nil {
				return nil, err
			}

			currentLocations = append(currentLocations, location)

			// delete MemoryUsageIterator if no new values
			if !v.Next() {
				delete(i.memoryUsageIteratorMap, l)
			} else if err := v.Error(); err != nil {
				return nil, err
			}
		}
	}

	return currentLocations, nil
}

func (i *measurementIterator) getLocation(callStack string, memUsage *schema.MemoryUsage) (*schema.Location, error) {
	cs := &schema.Callstack{}

	err := i.codec.decode(callStack, cs)
	if err != nil {
		return nil, err
	}

	return &schema.Location{
		MemoryUsage: memUsage,
		Callstack:   cs,
	}, nil
}

// NewMeasurementIterator iterator over measurements in session
func NewMeasurementIterator(tsdb prometheus.TSDB, codec codec, sessionLabel labels.Label) (MeasurementIterator, error) {
	querier, err := tsdb.Querier(0, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	// TODO: use LabelValuesFor when implemented
	// metaLabels, err := querier.LabelValuesFor(metaLabelName, sessionLabel)
	metaLabels, err := querier.LabelValues(metaLabelName)
	if err != nil {
		return nil, err
	}

	locationsIterMap := make(map[string]MemoryUsageIterator, len(metaLabels))

	// create map (map[Session]LocationIterator)
	for _, m := range metaLabels {
		metaLabel := labels.Label{Name: metaLabelName, Value: m}
		mui, ok := NewMemoryUsageIterator(querier, sessionLabel, metaLabel)
		if ok {
			locationsIterMap[m] = mui
		}
	}
	li := &measurementIterator{
		querier:                querier,
		currentTime:            time.Now().Unix(),
		memoryUsageIteratorMap: locationsIterMap,
		codec:                  codec,
		error:                  nil,
	}
	return li, nil
}
