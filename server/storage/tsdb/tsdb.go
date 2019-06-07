package tsdb

import (
	"context"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/tsdb"
)

type localStorage interface {
	querier(ctx context.Context, mint, maxt int64) (tsdb.Querier, error)
	startTime() int64
	appender() tsdb.Appender
	close() error
}

type localStorageImpl struct {
	db              *tsdb.DB
	startTimeMargin int64
}

func (s *localStorageImpl) startTime() int64 {
	var startTime int64

	if len(s.db.Blocks()) > 0 {
		startTime = s.db.Blocks()[0].Meta().MinTime
	} else {
		startTime = time.Now().Unix() * 1000
	}

	return startTime
}

func (s *localStorageImpl) querier(ctx context.Context, mint, maxt int64) (tsdb.Querier, error) {
	return s.db.Querier(mint, maxt)
}

func (s *localStorageImpl) appender() tsdb.Appender {
	return s.db.Appender()
}

func (s *localStorageImpl) close() error {
	return s.db.Close()
}

func newLocalStorage(db *tsdb.DB) localStorage {
	return &localStorageImpl{
		db: db,
	}
}

func Open(path string, l log.Logger, r prometheus.Registerer, opts *Options) (*tsdb.DB, error) {
	if opts.MinBlockDuration > opts.MaxBlockDuration {
		opts.MaxBlockDuration = opts.MinBlockDuration
	}
	rngs := tsdb.ExponentialBlockRanges(int64(time.Duration(opts.MinBlockDuration).Seconds()*1000), 10, 3)

	for i, v := range rngs {
		if v > int64(time.Duration(opts.MaxBlockDuration).Seconds()*1000) {
			rngs = rngs[:i]
			break
		}
	}

	db, err := tsdb.Open(path, l, r, &tsdb.Options{
		WALSegmentSize:         int(opts.WALSegmentSize),
		RetentionDuration:      uint64(time.Duration(opts.RetentionDuration).Seconds() * 1000),
		MaxBytes:               int64(opts.MaxBytes),
		BlockRanges:            rngs,
		NoLockfile:             opts.NoLockfile,
		AllowOverlappingBlocks: opts.AllowOverlappingBlocks,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Options struct {
	MinBlockDuration       model.Duration
	MaxBlockDuration       model.Duration
	WALSegmentSize         units.Base2Bytes
	RetentionDuration      model.Duration
	MaxBytes               units.Base2Bytes
	NoLockfile             bool
	AllowOverlappingBlocks bool
}
