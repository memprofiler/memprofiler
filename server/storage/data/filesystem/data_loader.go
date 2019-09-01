package filesystem

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"sync"

	"github.com/memprofiler/memprofiler/server/storage/data"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
)

type defaultDataLoader struct {
	codec  codec
	sd     *schema.SessionDescription
	fd     *os.File
	logger *zerolog.Logger
	wg     *sync.WaitGroup
}

const (
	loadChanCapacity = 256
)

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *data.LoadResult, error) {

	// prepare buffered channel for results
	results := make(chan *data.LoadResult, loadChanCapacity)

	scanner := bufio.NewScanner(l.fd)
	scanner.Split(bufio.ScanLines)

	// scan records line by line
	go func() {
		defer close(results)
		for scanner.Scan() {
			if len(scanner.Bytes()) > 0 {
				select {
				case results <- l.loadMeasurement(scanner.Bytes()):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return results, nil
}

// loadMeasurement disk
func (l *defaultDataLoader) loadMeasurement(in []byte) *data.LoadResult {
	var receiver schema.Measurement
	err := l.codec.decode(bytes.NewReader(in), &receiver)
	return &data.LoadResult{Measurement: &receiver, Err: err}
}

func (l *defaultDataLoader) Close() error {
	defer func() {
		l.wg.Done()
	}()
	return l.fd.Close()
}

func newDataLoader(
	dataFilePath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger *zerolog.Logger,
	wg *sync.WaitGroup,
) (data.Loader, error) {

	// open file to load records
	fd, err := os.Open(dataFilePath)
	if err != nil {
		return nil, err
	}

	contextLogger := logger.With().Fields(map[string]interface{}{
		"service":    sessionDesc.InstanceDescription.ServiceName,
		"instance":   sessionDesc.InstanceDescription.InstanceName,
		"session_id": sessionDesc.Id,
	}).Logger()

	loader := &defaultDataLoader{
		sd:     sessionDesc,
		fd:     fd,
		codec:  codec,
		logger: &contextLogger,
		wg:     wg,
	}
	return loader, nil
}
