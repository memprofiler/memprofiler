package test

import (
	"fmt"
	"go/build"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	serverConfig "github.com/memprofiler/memprofiler/server/config"
	serverLauncher "github.com/memprofiler/memprofiler/server/launcher"

	reporterConfig "github.com/memprofiler/memprofiler/test/reporter/config"
	reporterLauncher "github.com/memprofiler/memprofiler/test/reporter/launcher"
)

type clean func()

func runServer(logger logrus.FieldLogger, cfgPath string, errChan chan<- error) (clean, error) {

	// parse base config
	cfg, err := serverConfig.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, err
	}

	// override data dir
	cfg.Storage.Filesystem.DataDir = fmt.Sprintf("/tmp/memprofiler_%v", time.Now().Format("20060102150405"))

	l, err := serverLauncher.New(logger.WithField("side", "server"), cfg, errChan)
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
	l, err := reporterLauncher.New(logger.WithField("side", "reporter"), cfg, errChan)
	if err != nil {
		return nil, err
	}

	return l.Quit, nil
}

func TestIntegration(t *testing.T) {

	projectPath := filepath.Join(build.Default.GOPATH, "src/github.com/memprofiler/memprofiler")

	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	errChan := make(chan error, 2)

	// run memprofiler server
	serverCfg := filepath.Join(projectPath, "server/config/example.yml")
	serverCleanup, err := runServer(logger, serverCfg, errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer serverCleanup()

	// run built-in memprofiler client
	reporterCfg := filepath.Join(projectPath, "test/reporter/config/linear_growth.yml")
	reporterCleanup, err := runReporter(logger, reporterCfg, errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer reporterCleanup()

	time.Sleep(3 * time.Second)
}
