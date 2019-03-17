package main

import (
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/server/backend"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/frontend"
	"github.com/memprofiler/memprofiler/server/locator"
	"github.com/memprofiler/memprofiler/utils"
)

func run(cfg *config.Config) error {

	var (
		err     error
		errChan = make(chan error, 2)
	)

	// run subsystems
	locator, err := locator.NewLocator(cfg)
	if err != nil {
		return err
	}
	logger := locator.Logger
	defer locator.Quit()

	// run long-running tasks
	ss, err := runServices(locator, cfg, errChan)
	if err != nil {
		return err
	}
	ss.start(logger)
	defer ss.stop(logger)

	utils.BlockOnSignal(logger, errChan)
	return nil
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
	errChan chan error,
) (services, error) {

	var (
		err error
		ss  = services(map[string]common.Service{})
	)

	// 1. GRPC Backend
	ss[labelBackend], err = backend.NewServer(cfg.Backend, locator, errChan)
	if err != nil {
		return nil, err
	}

	// 2. Frontend
	ss[labelFrontend], err = frontend.NewServer(cfg.Frontend, locator, errChan)
	if err != nil {
		return nil, err
	}

	return ss, nil
}
