package prometheus_tsdb

import (
	"os"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/tsdb/labels"
	"github.com/stretchr/testify/assert"
)

// TestSimpleStorage test write and read from storage
func TestSimpleStorage(t *testing.T) {
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

	// create storage
	storage, err := OpenTSDB(testDir, logger)
	if err != nil {
		assert.FailNowf(t, "can not open database: %v", err.Error())
	}
	// close and cleanup after test
	defer func() {
		err := storage.Close()
		assert.NoError(t, err)
		err = os.RemoveAll(testDir)
		assert.NoError(t, err)
	}()

	// write data for labelSet
	appender := storage.Appender()
	for i := 1; i <= len(data); i++ {
		_, err := appender.Add(labelSet, int64(i), data[int64(i)])
		assert.NoError(t, err)
	}
	err = appender.Commit()
	assert.NoError(t, err)

	// read data with label0 (i.e. labelSet[0])
	querier, err := storage.Querier(0, 4)
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

// TestTwoLabelSetStorage test write and read from storage with filtering and parallel
func TestTwoLabelSetStorage(t *testing.T) {
	var (
		testDir = "/tmp/test"
		writer  = log.NewSyncWriter(os.Stdout)
		logger  = log.NewLogfmtLogger(writer)

		labelSets = []labels.Labels{
			{
				{Name: "meta", Value: "labelSet1"},
				{Name: "test1", Value: "test1"},
				{Name: "test2", Value: "test2"},
			},
			{
				{Name: "meta", Value: "labelSet2"},
				{Name: "test1", Value: "test1"},
				{Name: "test2", Value: "test2"},
			},
		}
		dataSets = []map[int64]float64{
			{
				1: 100,
				2: 200,
				3: 300,
			},
			{
				1: 101,
				2: 202,
				3: 303,
			},
		}
	)

	// create storage
	storage, err := OpenTSDB(testDir, logger)
	if err != nil {
		assert.FailNowf(t, "can not open database: %v", err.Error())
	}
	// close and cleanup after test
	defer func() {
		err := storage.Close()
		assert.NoError(t, err)
		err = os.RemoveAll(testDir)
		assert.NoError(t, err)
	}()

	var wg sync.WaitGroup
	// write data for labelSets
	for i := 0; i < 2; i++ {
		var (
			dataSet  = dataSets[i]
			labelSet = labelSets[i]
		)
		wg.Add(1)
		go func() {
			appender := storage.Appender()
			defer wg.Done()
			for j := 1; j <= len(dataSet); j++ {
				_, err := appender.Add(labelSet, int64(j), dataSet[int64(j)])
				assert.NoError(t, err)
			}
			err = appender.Commit()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	for i := 0; i < 2; i++ {
		var (
			dataSet  = dataSets[i]
			labelSet = labelSets[i]
		)
		// read data with label0 (i.e. labelSet[0])
		// read from 0 to 4 time
		querier, err := storage.Querier(0, 4)

		wg.Add(1)
		go func() {
			defer wg.Done()
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
					val, ok := dataSet[writeTime]
					if !ok {
						assert.Fail(t, "unexpected data: %v - %v", writeTime, val)
					}
					delete(dataSet, writeTime)
				}
			}

			// validate that all data found
			assert.Equal(t, 0, len(dataSet), "Not all data was returned")
		}()
	}
	wg.Wait()
}
