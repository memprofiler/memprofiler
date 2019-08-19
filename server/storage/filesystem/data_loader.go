package filesystem

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
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

func (l *defaultDataLoader) Load(ctx context.Context) (<-chan *storage.LoadResult, error) {

	// prepare bufferized channel for results
	results := make(chan *storage.LoadResult, loadChanCapacity)

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
func (l *defaultDataLoader) loadMeasurement(data []byte) *storage.LoadResult {
	var receiver schema.Measurement
	err := l.codec.decode(bytes.NewReader(data), &receiver)
	return &storage.LoadResult{Measurement: &receiver, Err: err}
}

func (l *defaultDataLoader) Close() error {
	defer l.wg.Done()
	return l.fd.Close()
}

func newDataLoader(
	subdirPath string,
	sessionDesc *schema.SessionDescription,
	codec codec,
	logger *zerolog.Logger,
	wg *sync.WaitGroup,
) (storage.DataLoader, error) {

	// open file to load records
	filename := filepath.Join(subdirPath, "data")
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	contextLogger := logger.With().Fields(map[string]interface{}{
		"type":        sessionDesc.GetServiceType(),
		"instance":    sessionDesc.GetServiceInstance(),
		"sessionDesc": storage.SessionIDToString(sessionDesc.GetSessionId()),
		"measurement": filename,
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
