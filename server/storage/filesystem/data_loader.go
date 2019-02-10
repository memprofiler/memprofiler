package filesystem

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"sync"

	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

type defaultDataLoader struct {
	cache  cache
	codec  codec
	sd     *storage.SessionDescription
	fd     *os.File
	logger logrus.FieldLogger
	mmID   measurementID
	wg     *sync.WaitGroup
}

const (
	loadChanCapacity = 256
)

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {

	// prepare bufferized channel for results
	results := make(chan *storage.LoadResult, loadChanCapacity)

	scanner := bufio.NewScanner(l.fd)
	scanner.Split(bufio.ScanLines)

	// scan records line by line
	go func() {
		defer close(results)
		for scanner.Scan() {
			select {
			case results <- l.loadMeasurement(scanner.Bytes()):
			case <-ctx.Done():
				return
			}
			l.mmID++
		}
	}()

	return results, nil
}

// loadMeasurement loads data either from in-memory cache, either from disk
func (l *defaultDataLoader) loadMeasurement(data []byte) *storage.LoadResult {

	meta := &measurementMetadata{
		session: l.sd,
		mmID:    l.mmID,
	}

	// try to load data from cache of unmarshaled values
	if cached := l.loadMeasurementFromCache(meta); cached != nil {
		return &storage.LoadResult{Measurement: cached}
	}

	// if value is missing in cache, take it directly from disk
	var receiver schema.Measurement
	err := l.codec.decode(bytes.NewReader(data), &receiver)
	if err == nil {
		// put it into cache
		l.cache.put(meta, &receiver)
	}
	return &storage.LoadResult{Measurement: &receiver, Err: err}
}

func (l *defaultDataLoader) loadMeasurementFromCache(mmMeta *measurementMetadata) *schema.Measurement {
	if l.cache == nil {
		return nil
	}
	value, _ := l.cache.get(mmMeta)
	return value
}

func (l *defaultDataLoader) saveMeasurementToCache(mmMeta *measurementMetadata, mm *schema.Measurement) {
	if l.cache == nil {
		return
	}
	l.cache.put(mmMeta, mm)
}

func (l *defaultDataLoader) Close() error {
	defer l.wg.Done()
	return l.fd.Close()
}

func newDataLoader(
	subdirPath string,
	sd *storage.SessionDescription,
	cache cache,
	codec codec,
	logger logrus.FieldLogger,
	wg *sync.WaitGroup,
) (storage.DataLoader, error) {

	// open file to load records
	filename := filepath.Join(subdirPath, "data")
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	contextLogger := logger.WithFields(logrus.Fields{
		"type":               sd.ServiceDescription.GetType(),
		"instance":           sd.ServiceDescription.GetInstance(),
		"sessionDescription": sd.SessionID,
		"measurement":        filename,
	})

	loader := &defaultDataLoader{
		sd:     sd,
		fd:     fd,
		cache:  cache,
		codec:  codec,
		logger: contextLogger,
		wg:     wg,
	}
	return loader, nil
}
