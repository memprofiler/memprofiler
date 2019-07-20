package tsdb

import (
	"fmt"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"

	"github.com/memprofiler/memprofiler/schema"
)

// MemoryUsageIterator memory usage statistic iterator for single location
type MemoryUsageIterator interface {
	Next() bool
	At() (int64, *schema.MemoryUsage)
	Error() error
}

var _ MemoryUsageIterator = (*memoryUsageIterator)(nil)

type memoryUsageIterator struct {
	// state of iteration
	currentTime     int64
	currentMemUsage *schema.MemoryUsage
	error           error

	// iterator for each memory usage type
	allocObjectsIterator tsdb.SeriesIterator
	allocBytesIterator   tsdb.SeriesIterator
	freeObjectsIterator  tsdb.SeriesIterator
	freeBytesIterator    tsdb.SeriesIterator
}

// CurrentTime current time and memory usage for location
func (i *memoryUsageIterator) At() (int64, *schema.MemoryUsage) {
	return i.currentTime, i.currentMemUsage
}

// Next check for next element and update state
func (i *memoryUsageIterator) Next() bool {
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
			i.error = fmt.Errorf("time for measurement is incorrect")
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

// Error current time and memory usage for location
func (i *memoryUsageIterator) Error() error {
	return i.error
}

// NewMemoryUsageIterator iterator for simple Location
func NewMemoryUsageIterator(querier tsdb.Querier, sessionLabel, metaLabel labels.Label) (MemoryUsageIterator, bool) {
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

	mui := &memoryUsageIterator{
		allocObjectsIterator: allocObjectsIterator,
		allocBytesIterator:   allocBytesIterator,
		freeObjectsIterator:  freeObjectsIterator,
		freeBytesIterator:    freeBytesIterator,

		error: nil,
	}

	// initial call Next
	next := mui.Next()
	if !next {
		return nil, false
	}
	if err := mui.Error(); err != nil {
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
