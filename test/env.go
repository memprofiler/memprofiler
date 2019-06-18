package test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/memprofiler/memprofiler/schema"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/server/common"
	server_config "github.com/memprofiler/memprofiler/server/config"
	server_launcher "github.com/memprofiler/memprofiler/server/launcher"
	reporter_config "github.com/memprofiler/memprofiler/test/reporter/config"
	reporter_launcher "github.com/memprofiler/memprofiler/test/reporter/launcher"
)

// env creates new testing environment (server with single reporter)
type env struct {
	server      common.Service                   // Memprofiler server
	serverCfg   *server_config.Config            // Memprofiler server config
	client      schema.MemprofilerFrontendClient // Memprofiler frontend client
	clientConn  *grpc.ClientConn                 // Memrpofiler frontend client
	reporter    common.Service                   // Memprofiler backend client (sends memory usage reports to servers)
	reporterCfg *reporter_config.Config          // Memprofiler backend client config
	logger      logrus.FieldLogger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func (e *env) Start() {
	e.reporter.Start()
}

func (e *env) Stop() {
	e.logger.Debug("Stopping reporter")
	e.reporter.Stop()
	e.logger.Debug("Stopping frontend client connection")
	if err := e.clientConn.Close(); err != nil {
		e.logger.WithError(err).Error("failed to close client conn")
	}
	e.logger.Debug("Stopping server")
	e.server.Stop()
	e.logger.Debug("Terminating loops")
	e.cancel()
	e.wg.Wait()
}

func newEnv(projectPath, serverConfigPath string) (*env, error) {

	ctx, cancel := context.WithCancel(context.Background())

	var (
		err     error
		errChan = make(chan error, 16) // FIXME: hopefully large enough, but try to determine it better
		l       = &env{
			logger: newLogger(),
			ctx:    ctx,
			cancel: cancel,
			wg:     sync.WaitGroup{},
		}
	)

	// wait for fatal errors or context cancellation
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		select {
		case <-l.ctx.Done():
			return
		case err = <-errChan:
			if err != nil {
				l.logger.Fatal(err)
			}
		}
	}()

	// run memprofiler server
	l.logger.Info("Starting memprofiler server")
	l.server, l.serverCfg, err = runServer(l.logger, serverConfigPath, errChan)
	if err != nil {
		return nil, err
	}
	l.server.Start()

	// run memprofiler client
	l.logger.Info("Starting memprofiler client")
	l.clientConn, err = grpc.Dial(l.serverCfg.Frontend.ListenEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	l.client = schema.NewMemprofilerFrontendClient(l.clientConn)

	// run test reporter
	l.logger.Info("Starting testing reporter")
	reporterCfgPath := filepath.Join(projectPath, "test/reporter/config/linear_growth.yml")
	l.reporter, l.reporterCfg, err = runReporter(l.logger, reporterCfgPath, errChan)
	if err != nil {
		return nil, err
	}
	l.reporter.Start()

	return l, nil
}

func runServer(logger logrus.FieldLogger, cfgPath string, errChan chan<- error,
) (common.Service, *server_config.Config, error) {

	// parse base config
	cfg, err := server_config.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	// override data dir
	dataDir := fmt.Sprintf("/tmp/memprofiler_%v", time.Now().Format("20060102150405"))
	switch {
	case cfg.Storage.Filesystem != nil:
		cfg.Storage.Filesystem.DataDir = dataDir
	case cfg.Storage.TSDB != nil:
		cfg.Storage.TSDB.DataDir = dataDir
	}

	l, err := server_launcher.New(logger.WithField("side", "server"), cfg, errChan)
	if err != nil {
		return nil, nil, err
	}
	return l, cfg, nil
}

func runReporter(logger logrus.FieldLogger, cfgPath string, errChan chan<- error,
) (common.Service, *reporter_config.Config, error) {

	cfg, err := reporter_config.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	l, err := reporter_launcher.New(logger.WithField("side", "reporter"), cfg, errChan)
	if err != nil {
		return nil, nil, err
	}

	return l, cfg, nil
}

func newLogger() logrus.FieldLogger {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	return logger
}
