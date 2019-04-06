package launcher

import (
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/server/backend"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/frontend"
	"github.com/memprofiler/memprofiler/server/locator"
)

type Launcher struct {
	locator  *locator.Locator
	logger   logrus.FieldLogger
	cfg      *config.Config
	services services
	errChan  chan<- error
}

func (l *Launcher) Start() {
	var err error
	l.services, err = runServices(l.locator, l.cfg, l.errChan)
	if err != nil {
		l.errChan <- err
	}
	l.services.start(l.logger)
}

func (l *Launcher) Stop() {
	l.locator.Quit()
	l.services.stop(l.locator.Logger)
}

func New(logger logrus.FieldLogger, cfg *config.Config, errChan chan<- error) (*Launcher, error) {

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

func (ss services) start(logger logrus.FieldLogger) {
	for label, service := range ss {
		logger.WithField("service", label).Info("Starting service")
		go service.Start()
	}
}

func (ss services) stop(logger logrus.FieldLogger) {
	for label, service := range ss {
		logger.WithField("service", label).Info("Stopping service")
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
	locator.Logger.Debug("Starting backend server")
	ss[labelBackend], err = backend.NewServer(cfg.Backend, locator, errChan)
	if err != nil {
		return nil, err
	}

	// 2. Frontend
	locator.Logger.Debug("Starting frontend server")
	ss[labelFrontend], err = frontend.NewServer(cfg.Frontend, locator, errChan)
	if err != nil {
		return nil, err
	}

	return ss, nil
}
