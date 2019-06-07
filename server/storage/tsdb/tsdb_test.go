package tsdb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/tsdb/labels"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalStorageReadWrite(t *testing.T) {
	var (
		testDir   = "/tmp/test"
		labelSet1 = labels.Labels{
			{Name: "test", Value: "test"},
			{Name: "test1", Value: "test1"},
			{Name: "test2", Value: "test2"},
			{Name: "test3", Value: "test3"},
		}
		labelSet2 = labels.Labels{
			{Name: "test", Value: "test"},
			{Name: "test4", Value: "test4"},
			{Name: "test5", Value: "test5"},
			{Name: "test6", Value: "test6"},
		}
		data1 = map[int64]float64{1000: 100, 2000: 200, 3000: 300}
		data2 = map[int64]float64{1000: 100, 2000: 200, 3000: 300}
	)

	// create logger
	writer := log.NewSyncWriter(os.Stdout)
	logger := log.NewLogfmtLogger(writer)

	// create db
	db, err := Open(testDir, logger, nil, &Options{
		// copy-past from tsdb examples
		MinBlockDuration: model.Duration(24 * time.Hour),
		MaxBlockDuration: model.Duration(24 * time.Hour),
	})
	if err != nil {
		assert.FailNowf(t, "can not open database: %v", err.Error())
	}

	defer func() {
		err := db.Close()
		assert.NoError(t, err)
		err = os.RemoveAll(testDir)
		assert.NoError(t, err)
	}()

	localStore := newLocalStorage(db)
	app := localStore.appender()

	// write data1 for labelSet1
	for k, v := range data1 {
		l, err := app.Add(labelSet1, k, v)
		assert.Equal(t, uint64(1), l)
		assert.NoError(t, err)
	}

	// write data2 for labelSet2
	for k, v := range data2 {
		_, err := app.Add(labelSet2, k, v)
		//assert.Equal(t, uint64(1), l)
		assert.NoError(t, err)
	}

	err = app.Commit()
	assert.NoError(t, err)

	// read data
	querier, err := localStore.querier(context.Background(), 1100, 3000)
	assert.NoError(t, err)
	delete(data1, 10)
	delete(data2, 10)

	//lNames, err := querier.LabelNames()
	//assert.NoError(t, err)
	//
	//for _, l := range labelSet1 {
	//	assert.True(t, hasStringVal(l.Name, lNames))
	//
	//	lVals, err := querier.LabelValues(l.Name)
	//	assert.NoError(t, err)
	//	assert.Equal(t, lVals[0], l.Name)
	//}
	//
	//for _, l := range labelSet2 {
	//	assert.True(t, hasStringVal(l.Name, lNames))
	//
	//	lVals, err := querier.LabelValues(l.Name)
	//	assert.NoError(t, err)
	//	assert.Equal(t, lVals[0], l.Name)
	//}

	seriesSet, _ := querier.Select([]labels.Matcher{
		labels.NewEqualMatcher("test", "test"),
	}...)

	for seriesSet.Next() {
		series := seriesSet.At()
		fmt.Printf("%v\n", series.Labels())
		seriesIterator := series.Iterator()
		for seriesIterator.Next() {
			writeTime, val := seriesIterator.At()
			fmt.Printf("%v-%v\n", writeTime, val)
			//val, ok := data[writeTime]
			//if !ok {
			//	assert.Fail(t, "unexpected data: %v - %v", writeTime, val)
			//}
		}
	}
}

func hasStringVal(val string, vals []string) bool {
	for _, v := range vals {
		if val == v {
			return true
		}
	}
	return false
}
