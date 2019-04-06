package test

import (
	"context"
	"fmt"
	"go/build"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/server/common"
	server_config "github.com/memprofiler/memprofiler/server/config"
	server_launcher "github.com/memprofiler/memprofiler/server/launcher"
	reporter_config "github.com/memprofiler/memprofiler/test/reporter/config"
	reporter_launcher "github.com/memprofiler/memprofiler/test/reporter/launcher"
)

// launcher creates new testing environment - a server with a single reporter
type launcher struct {
	server   common.Service
	reporter common.Service
	logger   logrus.FieldLogger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func (l *launcher) Start() {
	l.reporter.Start()
}

func (l *launcher) Stop() {
	l.reporter.Stop()
	l.server.Stop()
	l.cancel()
	l.wg.Wait()
}

func newLauncher() (common.Service, error) {
	const services = 2 // server + reporter = 2

	ctx, cancel := context.WithCancel(context.Background())

	var (
		errChan = make(chan error, services)
		err     error
		l       = &launcher{
			logger: newLogger(),
			ctx:    ctx,
			cancel: cancel,
			wg:     sync.WaitGroup{},
		}
	)

	// run memprofiler server
	l.logger.Info("Starting memprofiler server")
	projectPath := filepath.Join(build.Default.GOPATH, "src/github.com/memprofiler/memprofiler")
	serverCfg := filepath.Join(projectPath, "server/config/example.yml")
	l.server, err = runServer(l.logger, serverCfg, errChan)
	if err != nil {
		return nil, err
	}
	l.server.Start()

	// run test reporter
	l.logger.Info("Starting testing reporter")
	reporterCfg := filepath.Join(projectPath, "test/reporter/config/linear_growth.yml")
	l.reporter, err = runReporter(l.logger, reporterCfg, errChan)
	if err != nil {
		return nil, err
	}
	l.reporter.Start()

	// wait for fatal errors or context cancellation
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for i := 0; i < services; i++ {
			select {
			case <-l.ctx.Done():
				return
			case err = <-errChan:
				if err != nil {
					l.logger.Fatal(err)
				}
			}
		}
	}()

	return l, nil
}

func runServer(logger logrus.FieldLogger, cfgPath string, errChan chan<- error) (common.Service, error) {

	// parse base config
	cfg, err := server_config.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, err
	}

	// override data dir
	cfg.Storage.Filesystem.DataDir = fmt.Sprintf("/tmp/memprofiler_%v", time.Now().Format("20060102150405"))

	l, err := server_launcher.New(logger.WithField("side", "server"), cfg, errChan)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func runReporter(logger logrus.FieldLogger, cfgPath string, errChan chan<- error) (common.Service, error) {

	cfg, err := reporter_config.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, err
	}

	l, err := reporter_launcher.New(logger.WithField("side", "reporter"), cfg, errChan)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func newLogger() logrus.FieldLogger {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	return logger
}
