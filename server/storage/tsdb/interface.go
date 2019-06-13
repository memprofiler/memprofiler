package tsdb

import (
	"io"

	"github.com/prometheus/tsdb"
)

// Storage wrap tsdb as a point for possible expansion
type Storage interface {
	io.Closer

	// Appender adds data to storage
	Appender() tsdb.Appender
	// Querier query data from storage
	Querier(mint, maxt int64) (tsdb.Querier, error)

	// StartTime return lowest time in storage
	StartTime() int64
}
