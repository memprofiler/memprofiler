package metrics

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ Computer = (*defaultComputer)(nil)

type defaultComputer struct {
	mutex    sync.RWMutex
	sessions map[string]*sessionData
	logger   logrus.FieldLogger
	storage  storage.Storage
	cfg      *config.MetricsConfig
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// PutMeasurement stores measurement within internal storage
func (r *defaultComputer) PutMeasurement(sd *storage.SessionDescription, mm *schema.Measurement) error {
	r.mutex.Lock()
	data, exists := r.sessions[sd.String()]
	if !exists {
		data = newSessionData(r.logger, r.cfg.Window)
		r.sessions[sd.String()] = data
	}
	r.mutex.Unlock()

	data.registerMeasurement(mm)
	return nil
}

// GetSessionMetrics extracts the most recent metrics of a particular session
func (r *defaultComputer) GetSessionMetrics(
	ctx context.Context,
	sd *storage.SessionDescription,
) (*schema.SessionMetrics, error) {

	// get or create session data
	r.mutex.Lock()
	data, exists := r.sessions[sd.String()]
	if !exists {
		data = newSessionData(r.logger, r.cfg.Window)
		r.sessions[sd.String()] = data
	}
	r.mutex.Unlock()

	// if metrics of outdated session has been requested,
	// load data from storage and compute metrics from scratch
	if !exists {
		dataLoader, err := r.storage.NewDataLoader(sd)
		if err != nil {
			return nil, err
		}

		r.wg.Add(1)
		defer func() {
			if err = dataLoader.Close(); err != nil {
				r.logger.WithError(err).Error("Failed to close data loader")
			}
			r.wg.Done()
		}()

		loadChan, err := dataLoader.Load(ctx)
		if err != nil {
			return nil, err
		}

		if err := data.populate(ctx, loadChan); err != nil {
			return nil, err
		}
	}

	return data.getSessionMetrics(), nil
}

func (r *defaultComputer) populateSessionData(
	ctx context.Context,
	sd *storage.SessionDescription,
) (*schema.SessionMetrics, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// prepare new session data storage
	r.mutex.Lock()
	data := newSessionData(r.logger, r.cfg.Window)
	r.sessions[sd.String()] = data
	r.mutex.Unlock()

	return data.getSessionMetrics(), nil
}

//sort.Slice(locations, func(i, j int) bool {
//	// descending order
//	return locations[i].Average.InUseBytesRate > locations[j].Average.InUseBytesRate
//})

func (r *defaultComputer) Quit() {
	r.cancel()
	r.wg.Wait()
}

// NewComputer instantiates new runner
func NewComputer(logger logrus.FieldLogger, storage storage.Storage, cfg *config.MetricsConfig) Computer {
	ctx, cancel := context.WithCancel(context.Background())
	return &defaultComputer{
		logger:   logger,
		sessions: make(map[string]*sessionData),
		mutex:    sync.RWMutex{},
		wg:       sync.WaitGroup{},
		storage:  storage,
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
	}
}
