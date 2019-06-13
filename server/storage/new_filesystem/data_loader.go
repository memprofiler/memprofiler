package filesystem

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
	localTSDB "github.com/memprofiler/memprofiler/server/storage/tsdb"
)

type defaultDataLoader struct {
	storage localTSDB.Storage
	codec   codec
	sd      *schema.SessionDescription
	logger  logrus.FieldLogger
	wg      *sync.WaitGroup
}

const (
	loadChanCapacity = 256
)

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {
	var (
		sessionLabel = labels.Label{Name: SessionLabelName, Value: fmt.Sprintf("%v", l.sd.GetSessionId())}
	)

	li, err := NewLocationsIter(l.storage, l.codec, sessionLabel)
	if err != nil {
		return nil, err
	}

	// prepare bufferized channel for results
	results := make(chan *storage.LoadResult, loadChanCapacity)
	go func() {
		for li.Next() {
			m := &storage.LoadResult{Measurement: li.At(), Err: err}
			select {
			case results <- m:
			case <-ctx.Done():
				break
			}
		}
		close(results)
	}()

	return results, nil
}

func (l *defaultDataLoader) Close() error {
	defer l.wg.Done()
	return nil
}

func newDataLoader(
	subdirPath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger logrus.FieldLogger,
	wg *sync.WaitGroup,
) (storage.DataLoader, error) {

	// open file to load records
	contextLogger := logger.WithFields(logrus.Fields{
		"type":        sessionDesc.GetServiceType(),
		"instance":    sessionDesc.GetServiceInstance(),
		"sessionDesc": storage.SessionIDToString(sessionDesc.GetSessionId()),
	})
	var (
		writer  = log.NewSyncWriter(os.Stdout)
		logger2 = log.NewLogfmtLogger(writer)
	)

	// create storage
	stor, err := localTSDB.OpenStorage(subdirPath, logger2)
	if err != nil {
		return nil, err
	}

	loader := &defaultDataLoader{
		storage: stor,
		sd:      sessionDesc,
		codec:   codec,
		logger:  contextLogger,
		wg:      wg,
	}
	return loader, nil
}

// LocationsIter тут бегаем по локациям и достаем наборы
type LocationsIter struct {
	querier     tsdb.Querier
	currentTime int64
	codec       codec

	locationsIterMap map[string]*MemoryUsageIterator
}

func NewLocationsIter(tsdb localTSDB.Storage, codec codec, sessionLabel labels.Label) (*LocationsIter, error) {
	querier, err := tsdb.Querier(0, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	metaLabels, err := querier.LabelValues(MetaLabelName)
	if err != nil {
		return nil, err
	}

	locationsIterMap := make(map[string]*MemoryUsageIterator, len(metaLabels))

	for _, m := range metaLabels {
		metaLabel := labels.Label{Name: MetaLabelName, Value: m}
		locationsIterMap[m] = NewMemoryUsageIterator(querier, sessionLabel, metaLabel)
	}
	return &LocationsIter{
		querier:          querier,
		currentTime:      0,
		locationsIterMap: locationsIterMap,
		codec:            codec,
	}, nil
}

func (i *LocationsIter) Next() bool {
	if len(i.locationsIterMap) > 0 {
		return true
	}
	return false
}

func (i *LocationsIter) At() *schema.Measurement {
	m := &schema.Measurement{
		ObservedAt: &timestamp.Timestamp{Seconds: i.currentTime},
		Locations:  i.currentLocations(),
	}
	i.updateMin()
	return m
}

func (i *LocationsIter) updateMin() {
	for _, v := range i.locationsIterMap {
		if v.currentTime < i.currentTime {
			i.currentTime = v.currentTime
		}
	}
}

func (i *LocationsIter) currentLocations() []*schema.Location {
	var currentLocations []*schema.Location
	for l, v := range i.locationsIterMap {
		if v.currentTime == i.currentTime {
			currentLocations = append(currentLocations, i.getLocation(l, v.CurrentValue()))
			if !i.locationsIterMap[l].Next() {
				delete(i.locationsIterMap, l)
			}
		}
	}
	return currentLocations
}

func (i *LocationsIter) getLocation(callStack string, memUsage *schema.MemoryUsage) *schema.Location {
	cs := &schema.Callstack{}
	err := i.codec.decode(callStack, cs)
	if err != nil {
		panic(err)
	}
	return &schema.Location{
		MemoryUsage: memUsage,
		Callstack:   cs,
	}
}

// MemoryUsageIterator для каждой локации статистика
type MemoryUsageIterator struct {
	currentTime     int64
	currentMemUsage *schema.MemoryUsage

	allocObjectsIterator tsdb.SeriesIterator
	allocBytesIterator   tsdb.SeriesIterator
	freeObjectsIterator  tsdb.SeriesIterator
	freeBytesIterator    tsdb.SeriesIterator
}

func NewMemoryUsageIterator(querier tsdb.Querier, sessionLabel, metaLabel labels.Label) *MemoryUsageIterator {
	mui := &MemoryUsageIterator{
		allocObjectsIterator: getSeriesIterator(querier, sessionLabel, metaLabel, AllocObjectsLabel),
		allocBytesIterator:   getSeriesIterator(querier, sessionLabel, metaLabel, AllocBytesLabel),
		freeObjectsIterator:  getSeriesIterator(querier, sessionLabel, metaLabel, FreeObjectsLabel),
		freeBytesIterator:    getSeriesIterator(querier, sessionLabel, metaLabel, FreeBytesLabel),
	}
	mui.Next()
	return mui
}
func (i *MemoryUsageIterator) CurrentTime() int64 {
	return i.currentTime
}
func (i *MemoryUsageIterator) CurrentValue() *schema.MemoryUsage {
	return i.currentMemUsage
}

func (i *MemoryUsageIterator) Next() bool {
	if i.allocObjectsIterator.Next() &&
		i.allocBytesIterator.Next() &&
		i.freeObjectsIterator.Next() &&
		i.freeBytesIterator.Next() {

		t1, allocObjects := i.allocObjectsIterator.At()
		t2, allocBytes := i.allocBytesIterator.At()
		t3, freeObjects := i.freeObjectsIterator.At()
		t4, freeBytes := i.freeBytesIterator.At()

		if !(t1 == t2 && t2 == t3 && t3 == t4) {
			panic("HOW and WHY")
		}

		i.currentMemUsage = &schema.MemoryUsage{
			AllocObjects: int64(allocObjects),
			AllocBytes:   int64(allocBytes),
			FreeObjects:  int64(freeObjects),
			FreeBytes:    int64(freeBytes),
		}
		return true
	}
	return false
}
func getSeriesIterator(querier tsdb.Querier, sessionLabel, metaLabel, l labels.Label) tsdb.SeriesIterator {
	seriesSet, _ := querier.Select([]labels.Matcher{
		labels.NewEqualMatcher(sessionLabel.Name, sessionLabel.Value),
		labels.NewEqualMatcher(metaLabel.Name, metaLabel.Value),
		labels.NewEqualMatcher(l.Name, l.Value),
	}...)
	seriesSet.Next()
	series := seriesSet.At()
	return series.Iterator()
}
