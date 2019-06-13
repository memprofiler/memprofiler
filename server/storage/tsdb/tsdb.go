package tsdb

import (
	"github.com/go-kit/kit/log"
	"github.com/prometheus/tsdb"
)

var _ Storage = (*defaultStorage)(nil)

type defaultStorage struct {
	db *tsdb.DB
}

func (s *defaultStorage) Close() error {
	return s.db.Close()
}

func (s *defaultStorage) Appender() tsdb.Appender {
	return s.db.Appender()
}

func (s *defaultStorage) Querier(mint, maxt int64) (tsdb.Querier, error) {
	return s.db.Querier(mint, maxt)
}

func (s *defaultStorage) StartTime() int64 {
	var startTime int64

	if len(s.db.Blocks()) > 0 {
		startTime = s.db.Blocks()[0].Meta().MinTime
	} else {
		startTime = s.db.Head().MinTime()
	}

	return startTime
}

func OpenStorage(dir string, l log.Logger) (Storage, error) {
	db, err := tsdb.Open(dir, l, nil, tsdb.DefaultOptions)
	if err != nil {
		return nil, err
	}
	return &defaultStorage{
		db: db,
	}, nil
}
