package test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"

	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	serverConfig "github.com/memprofiler/memprofiler/server/config"
	serverLauncher "github.com/memprofiler/memprofiler/server/launcher"
	reporterConfig "github.com/memprofiler/memprofiler/test/reporter/config"
	reporterLauncher "github.com/memprofiler/memprofiler/test/reporter/launcher"
)

// env creates new testing environment (server with single reporter)
type env struct {
	server      common.Service                   // Memprofiler server
	serverCfg   *serverConfig.Config             // Memprofiler server config
	client      schema.MemprofilerFrontendClient // Memprofiler frontend client
	clientConn  *grpc.ClientConn                 // Memprofiler frontend client
	reporter    common.Service                   // Memprofiler backend client (sends memory usage reports to servers)
	reporterCfg *reporterConfig.Config           // Memprofiler backend client config
	logger      *zerolog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func (e *env) Start() {
	e.reporter.Start()
}

func (e *env) Stop() {
	e.logger.Debug().Msg("Stopping reporter")
	e.reporter.Stop()
	e.logger.Debug().Msg("Stopping frontend client connection")
	if err := e.clientConn.Close(); err != nil {
		e.logger.Err(err).Msg("failed to close client conn")
	}
	e.logger.Debug().Msg("Stopping server")
	e.server.Stop()
	e.logger.Debug().Msg("Terminating loops")
	e.cancel()
	e.wg.Wait()
}

func newEnv(projectPath, serverConfigPath string) (*env, error) {

	ctx, cancel := context.WithCancel(context.Background())

	var (
		err     error
		errChan = make(chan error, 16) // FIXME: hopefully large enough, but try to determine it better
		l       = &env{
			logger: utils.NewLogger(&config.LoggingConfig{Level: zerolog.DebugLevel}),
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
				l.logger.Fatal().Err(err).Send()
			}
		}
	}()

	// run memprofiler server
	l.logger.Info().Msg("Starting memprofiler server")
	l.server, l.serverCfg, err = runServer(l.logger, serverConfigPath, errChan)
	if err != nil {
		return nil, err
	}
	l.server.Start()

	// run memprofiler client
	l.logger.Info().Msg("Starting memprofiler client")
	l.clientConn, err = grpc.Dial(l.serverCfg.Frontend.ListenEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	l.client = schema.NewMemprofilerFrontendClient(l.clientConn)

	// run test reporter
	l.logger.Info().Msg("Starting testing reporter")
	reporterCfgPath := filepath.Join(projectPath, "test/reporter/config/linear_growth.yml")
	l.reporter, l.reporterCfg, err = runReporter(l.logger, reporterCfgPath, errChan)
	if err != nil {
		return nil, err
	}
	l.reporter.Start()

	return l, nil
}

func runServer(
	logger *zerolog.Logger,
	cfgPath string,
	errChan chan<- error,
) (
	common.Service,
	*serverConfig.Config,
	error,
) {

	// parse base config
	cfg, err := serverConfig.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("open file %v", cfgPath))
	}

	// override data dir
	dataDir := fmt.Sprintf("/tmp/memprofiler_%v", time.Now().Format("20060102150405"))
	switch cfg.DataStorage.Type() {
	case config.FilesystemDataStorage:
		cfg.DataStorage.Filesystem.DataDir = dataDir
	case config.TSDBDataStorage:
		cfg.DataStorage.TSDB.DataDir = dataDir
	}

	subLogger := logger.With().Fields(map[string]interface{}{
		"side": "server",
	}).Logger()

	l, err := serverLauncher.New(&subLogger, cfg, errChan)
	if err != nil {
		return nil, nil, err
	}
	return l, cfg, nil
}

func runReporter(logger *zerolog.Logger, cfgPath string, errChan chan<- error,
) (common.Service, *reporterConfig.Config, error) {

	cfg, err := reporterConfig.FromYAMLFile(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	subLogger := logger.With().Fields(map[string]interface{}{
		"side": "reporter",
	}).Logger()

	l, err := reporterLauncher.New(&subLogger, cfg, errChan)
	if err != nil {
		return nil, nil, err
	}

	return l, cfg, nil
}
