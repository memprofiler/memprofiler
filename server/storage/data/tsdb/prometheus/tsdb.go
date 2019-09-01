package prometheus

import (
	"github.com/go-kit/kit/log"
	"github.com/prometheus/tsdb"
)

var _ TSDB = (*defaultTSDB)(nil)

type defaultTSDB struct {
	db *tsdb.DB
}

// Close close tsdb
func (s *defaultTSDB) Close() error {
	return s.db.Close()
}

// Appender adds data to storage
func (s *defaultTSDB) Appender() tsdb.Appender {
	return s.db.Appender()
}

// Querier query data from storage
func (s *defaultTSDB) Querier(mint, maxt int64) (tsdb.Querier, error) {
	return s.db.Querier(mint, maxt)
}

// StartTime return lowest time in storage
func (s *defaultTSDB) StartTime() int64 {
	var startTime int64

	if len(s.db.Blocks()) > 0 {
		startTime = s.db.Blocks()[0].Meta().MinTime
	} else {
		startTime = s.db.Head().MinTime()
	}

	return startTime
}

// OpenTSDB open tsdb in specified dir
func OpenTSDB(dir string, l log.Logger) (TSDB, error) {
	db, err := tsdb.Open(dir, l, nil, tsdb.DefaultOptions)
	if err != nil {
		return nil, err
	}

	return &defaultTSDB{
		db: db,
	}, nil
}
