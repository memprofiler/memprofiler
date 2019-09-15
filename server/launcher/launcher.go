package launcher

import (
	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/server/backend"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/frontend"
	"github.com/memprofiler/memprofiler/server/locator"
)

// Launcher is a top-level server structure responsible for Start and Stop
type Launcher struct {
	locator  *locator.Locator
	logger   *zerolog.Logger
	cfg      *config.Config
	services services
	errChan  chan<- error
}

// Start runs internal services
func (l *Launcher) Start() {
	var err error
	l.services, err = runServices(l.locator, l.cfg, l.errChan)
	if err != nil {
		l.errChan <- err
		return
	}
	l.services.start(l.logger)
}

// Stop stops internal services
func (l *Launcher) Stop() {
	l.locator.Quit()
	l.services.stop(l.locator.Logger)
}

// New prepares new launcher
func New(logger *zerolog.Logger, cfg *config.Config, errChan chan<- error) (*Launcher, error) {

	// run subsystems
	l, err := locator.NewLocator(logger, cfg)
	if err != nil {
		return nil, err
	}

	result := &Launcher{
		locator: l,
		logger:  logger,
		cfg:     cfg,
		errChan: errChan,
	}
	return result, nil
}

type services map[string]common.Service

func (ss services) start(logger *zerolog.Logger) {
	for label, service := range ss {
		logger.Info().Str("service", label).Msg("Starting service")
		go service.Start()
	}
}

func (ss services) stop(logger *zerolog.Logger) {
	for label, service := range ss {
		logger.Info().Str("service", label).Msg("Stopping service")
		service.Stop()
	}
}

const (
	labelBackend  = "backend"
	labelFrontend = "frontend"
)

func runServices(
	locator *locator.Locator,
	cfg *config.Config,
	errChan chan<- error,
) (services, error) {

	var (
		err error
		ss  = services(map[string]common.Service{})
	)

	// 1. GRPC Backend
	locator.Logger.Debug().Msg("Starting backend server")
	ss[labelBackend], err = backend.NewServer(cfg.Backend, locator, errChan)
	if err != nil {
		return nil, err
	}

	// 2. Frontend
	locator.Logger.Debug().Msg("Starting frontend server")
	ss[labelFrontend], err = frontend.NewServer(cfg.Frontend, locator, errChan)
	if err != nil {
		return nil, err
	}

	return ss, nil
}
