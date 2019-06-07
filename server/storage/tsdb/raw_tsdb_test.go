package tsdb

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
	"github.com/stretchr/testify/assert"
)

// TestRawTSDB test write and read from prometheus tsdb
func TestRawTSDB(t *testing.T) {
	var (
		testDir  = "/tmp/test"
		writer   = log.NewSyncWriter(os.Stdout)
		logger   = log.NewLogfmtLogger(writer)
		labelSet = labels.Labels{
			{Name: "test0", Value: "test0"},
			{Name: "test1", Value: "test1"},
			{Name: "test2", Value: "test2"},
		}
		data = map[int64]float64{
			1: 100,
			2: 200,
			3: 300,
		}
	)

	// create db
	db, err := tsdb.Open(testDir, logger, nil, tsdb.DefaultOptions)
	if err != nil {
		assert.FailNowf(t, "can not open database: %v", err.Error())
	}
	// close and cleanup after test
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
		err = os.RemoveAll(testDir)
		assert.NoError(t, err)
	}()

	// write data for labelSet
	appender := db.Appender()
	for i := 1; i <= len(data); i++ {
		_, err := appender.Add(labelSet, int64(i), data[int64(i)])
		assert.NoError(t, err)
	}
	err = appender.Commit()
	assert.NoError(t, err)

	// read data with label0 (i.e. labelSet[0])
	querier, err := db.Querier(0, 4)
	assert.NoError(t, err)
	seriesSet, err := querier.Select([]labels.Matcher{
		labels.NewEqualMatcher(labelSet[0].Name, labelSet[0].Value),
	}...)
	assert.NoError(t, err)
	for seriesSet.Next() {
		series := seriesSet.At()
		seriesIterator := series.Iterator()
		for seriesIterator.Next() {
			writeTime, val := seriesIterator.At()
			// validate data, if data exist, delete it
			val, ok := data[writeTime]
			if !ok {
				assert.Fail(t, "unexpected data: %v - %v", writeTime, val)
			}
			delete(data, writeTime)
		}
	}

	// validate that all data found
	assert.Equal(t, 0, len(data), "Not all data was returned")
}
