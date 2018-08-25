package filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"sync"

	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

type defaultDataLoader struct {
	cache      cache
	codec      codec
	sd         *storage.SessionDescription
	subdirPath string
	logger     logrus.FieldLogger
	wg         *sync.WaitGroup
}

const (
	maxLoadChanCapacity = 256
)

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {

	files, err := ioutil.ReadDir(l.subdirPath)
	if err != nil {
		return nil, err
	}

	// prepare bufferized channel for results
	var loadChanCapacity int
	if len(files) > maxLoadChanCapacity {
		loadChanCapacity = maxLoadChanCapacity
	} else {
		loadChanCapacity = len(files)
	}
	results := make(chan *storage.LoadResult, loadChanCapacity)

	// take files from disk, deserialize it and send it to the caller asynchronously
	go func() {
		defer close(results)
		for _, file := range files {
			select {
			case results <- l.loadMeasurement(file.Name()):
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

// loadMeasurement loads data either from in-memory cache, either from disk
func (l *defaultDataLoader) loadMeasurement(filename string) *storage.LoadResult {

	contextLogger := l.logger.WithFields(logrus.Fields{
		"type":        l.sd.ServiceDescription.GetType(),
		"instance":    l.sd.ServiceDescription.GetInstance(),
		"session":     l.sd.SessionID,
		"measurement": filename,
	})

	mmMeta, err := l.makeMeasurementMetadata(filename)
	if err != nil {
		return &storage.LoadResult{Err: err}
	}

	// try to load data from cache of unmarshaled values
	if cached := l.loadMeasurementFromCache(mmMeta); cached != nil {
		contextLogger.Debug("Loaded from cache")
		return &storage.LoadResult{Measurement: cached}
	}

	// if value is missing in cache, take it directly from disk
	mm, err := l.loadMeasurementFromFile(filename)
	if err == nil {
		contextLogger.Debug("Loaded from disk")

		// put it into cache
		l.cache.put(mmMeta, mm)
	}
	return &storage.LoadResult{Measurement: mm, Err: err}
}

func (l *defaultDataLoader) makeMeasurementMetadata(filename string) (*measurementMetadata, error) {
	// convert filename to measurement ID
	mmID, err := measurementIDFromString(filename)
	if err != nil {
		return nil, err
	}

	// try to load data from cache
	mmMeta := &measurementMetadata{
		SessionDescription: *l.sd,
		mmID:               mmID,
	}
	return mmMeta, nil
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

// loadMeasurementFromFile reads single measurement from file
func (l *defaultDataLoader) loadMeasurementFromFile(filename string) (*schema.Measurement, error) {
	path := filepath.Join(l.subdirPath, filename)

	fd, err := os.OpenFile(path, os.O_RDONLY, filePermissions)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var measurement schema.Measurement
	if err := l.codec.decode(fd, &measurement); err != nil {
		return nil, err
	}

	return &measurement, nil
}

func (l *defaultDataLoader) Close() error {
	l.wg.Done()
	return nil
}
