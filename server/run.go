package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/server/backend"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/frontend"
	"github.com/memprofiler/memprofiler/server/locator"
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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case <-signalChan:
		logger.Warning("Interrupt signal has been received")
	case err := <-errChan:
		if err != nil {
			logger.WithError(err).Error("Fatal error, going to terminate server")
		} else {
			logger.Warning("Going to terminate server")
		}
	}

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
	labelAPI = "api"
	labelWeb = "web"
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

	// 1. GRPC API
	ss[labelAPI], err = backend.NewServer(cfg.API, locator, errChan)
	if err != nil {
		return nil, err
	}

	// 2. Web
	ss[labelWeb], err = frontend.NewServer(cfg.Web, locator, errChan)
	if err != nil {
		return nil, err
	}

	return ss, nil
}
