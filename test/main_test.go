package test

import (
	"testing"

	"github.com/sirupsen/logrus"

	serverConfig "github.com/memprofiler/memprofiler/server/config"
	serverLauncher "github.com/memprofiler/memprofiler/server/launcher"

	reporterConfig "github.com/memprofiler/memprofiler/test/reporter/config"
	reporterLauncher "github.com/memprofiler/memprofiler/test/reporter/launcher"
)

type clean func()

func runServer(logger logrus.FieldLogger, cfgPath string, errChan chan<- error) (clean, error) {

	cfg, err := serverConfig.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, err
	}

	l, err := serverLauncher.New(logger, cfg, errChan)
	if err != nil {
		return nil, err
	}
	l.Start()
	return l.Stop, nil
}

func runReporter(logger logrus.FieldLogger, cfgPath string, errChan chan<- error) (clean, error) {

	cfg, err := reporterConfig.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, err
	}
	l, err := reporterLauncher.New(logger, cfg, errChan)
	if err != nil {
		return nil, err
	}

	return l.Quit, nil
}

func TestIntegration(t *testing.T) {

	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	errChan := make(chan error, 2)

	// run memprofiler server
	serverCleanup, err := runServer(logger, "./server/config/example.yml", errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer serverCleanup()

	// run built-in memprofiler client
	reporterCleanup, err := runReporter(logger, "./test/reporter/config/linear_growth.yml", errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer reporterCleanup()
}
