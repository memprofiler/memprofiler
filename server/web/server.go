package web

import (
	"context"
	"net/http"

	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/locator"
	"github.com/vitalyisaev2/memprofiler/server/metrics"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

type server struct {
	httpServer *http.Server
	computer   metrics.Computer
	storage    storage.Storage
	errChan    chan<- error
	logger     logrus.FieldLogger
}

func (*server) Save(schema.Memprofiler_SaveServer) error {
	panic("implement me")
}

// Start runs HTTP API
func (s *server) Start() { s.errChan <- s.httpServer.ListenAndServe() }

const (
	terminationTimeout = 1 * time.Second
)

// Stop terminates HTTP API
func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), terminationTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil && err != ctx.Err() {
		s.logger.WithError(err).Error("API shutdown error")
	}
}

// NewServer initializes new server
func NewServer(
	cfg *config.WebConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (common.Service, error) {

	s := &server{
		httpServer: &http.Server{
			Addr: cfg.ListenEndpoint,
		},
		computer: locator.Computer,
		storage:  locator.Storage,
		logger:   locator.Logger,
		errChan:  errChan,
	}

	r := httprouter.New()
	r.GET("/computer/:type/:instance/:session", s.computeSessionMetrics)

	s.httpServer.Handler = r

	return s, nil
}
