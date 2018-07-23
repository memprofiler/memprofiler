package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/server/api"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/locator"
)

const (
	labelAPI = "api"
	labelWeb = "web"
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

func (ss services) stop(logger *logrus.Logger) {
	for label, service := range ss {
		logger.WithField("service", label).Info("Stopping service")
		service.Stop()
	}
}

func runServices(
	locator *locator.Locator,
	cfg *config.Config,
	errChan chan error,
) (services, error) {

	var (
		err    error
		ss     = services(map[string]common.Service{})
		logger = locator.Logger
	)

	// 1. GRPC API
	logger.WithField("service", labelAPI).Info("Starting service")
	ss[labelAPI], err = api.NewAPI(cfg.Server, locator, errChan)
	if err != nil {
		return nil, err
	}

	return ss, nil
}
