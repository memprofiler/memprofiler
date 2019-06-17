package prometheus_tsdb

import (
	"io"

	"github.com/prometheus/tsdb"
)

// TSDB wrap prometheus tsdb as a point for possible expansion
type TSDB interface {
	io.Closer

	// Appender adds data to storage
	Appender() tsdb.Appender
	// Querier query data from storage
	Querier(mint, maxt int64) (tsdb.Querier, error)

	// StartTime return lowest time in storage
	StartTime() int64
}
