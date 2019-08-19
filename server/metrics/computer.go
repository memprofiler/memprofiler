package metrics

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage"
)

var _ Computer = (*defaultComputer)(nil)

type defaultComputer struct {
	// sessions contains time series "tails" with the most recent session data.
	// This data is used to recompute trend values as soon as new measurements come.
	// FIXME: it's necessary to implement session cleanup, otherwise memory will leak
	sessions map[string]*sessionData

	// dispatcher owns subscriptions
	dispatcher dispatcher

	// storage provides data that is not in cache yet
	storage storage.Storage

	mutex  sync.RWMutex
	cfg    *config.MetricsConfig
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	logger *zerolog.Logger
}

// PutMeasurement stores measurement within internal storage
func (r *defaultComputer) PutMeasurement(sd *schema.SessionDescription, mm *schema.Measurement) error {
	r.mutex.Lock()
	sessionID := shortSessionIdentifier(sd)
	data, exists := r.sessions[sessionID]
	if !exists {
		data = newSessionData(r.logger, r.cfg.AveragingWindows)
		r.sessions[sessionID] = data
	}
	r.mutex.Unlock()

	// push measurement to time series
	if err := data.appendMeasurement(mm); err != nil {
		return err
	}

	// notify subscribers
	r.dispatcher.broadcast(sd, data.getSessionMetrics())
	return nil
}

// SessionRecentMetrics extracts the most recent metrics of a particular session
func (r *defaultComputer) SessionRecentMetrics(
	ctx context.Context,
	sd *schema.SessionDescription,
) (*schema.SessionMetrics, error) {

	sessionID := shortSessionIdentifier(sd)

	// get or create session data
	r.mutex.Lock()
	data, exists := r.sessions[sessionID]
	if !exists {
		data = newSessionData(r.logger, r.cfg.AveragingWindows)
		r.sessions[sessionID] = data
	}
	r.mutex.Unlock()

	// if metrics of old session has been requested, that doesn't exist in cache yet,
	// load data from storage and compute metrics from scratch
	if !exists {
		if err := r.populateSessionData(ctx, sd, data); err != nil {
			return nil, err
		}
	}

	return data.getSessionMetrics(), nil
}

func (r *defaultComputer) SessionSubscribe(ctx context.Context, sd *schema.SessionDescription) (Subscription, error) {

	sessionID := shortSessionIdentifier(sd)

	// check session data, create if not exists
	r.mutex.Lock()
	data, exists := r.sessions[sessionID]
	if !exists {
		data = newSessionData(r.logger, r.cfg.AveragingWindows)
		r.sessions[sessionID] = data
	}
	r.mutex.Unlock()

	// if metrics of old session has been requested, that doesn't exist in cache yet,
	// load data from storage and compute metrics from scratch
	if !exists {
		if err := r.populateSessionData(ctx, sd, data); err != nil {
			return nil, err
		}
	}

	subscription := r.dispatcher.createSubscription(ctx, sd)
	r.dispatcher.broadcast(sd, data.getSessionMetrics())
	return subscription, nil
}

// populateSessionData takes data from persistent storage
func (r *defaultComputer) populateSessionData(
	ctx context.Context,
	sd *schema.SessionDescription,
	data *sessionData) error {

	// TODO: don't load all data, take only tail needed by the largest averaging window;
	// need to extend storage interface
	dataLoader, err := r.storage.NewDataLoader(sd)
	if err != nil {
		return err
	}

	r.wg.Add(1)
	defer func() {
		if err = dataLoader.Close(); err != nil {
			r.logger.Err(err).Msg("Failed to close data loader")
		}
		r.wg.Done()
	}()

	loadChan, err := dataLoader.Load(ctx)
	if err != nil {
		return err
	}

	return data.populate(ctx, loadChan)
}

func (r *defaultComputer) Quit() {
	r.cancel()
	r.wg.Wait()
}

// NewComputer instantiates new runner
func NewComputer(logger *zerolog.Logger, storage storage.Storage, cfg *config.MetricsConfig) Computer {
	ctx, cancel := context.WithCancel(context.Background())
	return &defaultComputer{
		logger:     logger,
		sessions:   make(map[string]*sessionData),
		dispatcher: newDispatcher(),
		storage:    storage,
		mutex:      sync.RWMutex{},
		wg:         sync.WaitGroup{},
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
	}
}
