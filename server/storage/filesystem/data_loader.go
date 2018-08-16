package filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"sync"

	"path/filepath"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

type defaultDataLoader struct {
	codec      codec
	subdirPath string
	sessionID  storage.SessionID
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
			path := filepath.Join(l.subdirPath, file.Name())
			select {
			case results <- l.loadSingleMeasurement(path):
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

// loadSingleMeasurement reads single measurement from file
func (l *defaultDataLoader) loadSingleMeasurement(filePath string) (result *storage.LoadResult) {
	result = &storage.LoadResult{}

	fd, err := os.OpenFile(filePath, os.O_RDONLY, filePermissions)
	if err != nil {
		result.Err = err
		return
	}
	defer fd.Close()

	var measurement schema.Measurement
	if err := l.codec.decode(fd, &measurement); err != nil {
		result.Err = err
		return
	}

	result.Measurement = &measurement
	return
}

func (l *defaultDataLoader) Close() error {
	l.wg.Done()
	return nil
}
