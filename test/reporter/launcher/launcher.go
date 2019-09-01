package launcher

import (
	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/test/reporter/config"
	"github.com/memprofiler/memprofiler/test/reporter/playback"
)

type launcher struct {
	profiler client.Profiler
	playback playback.Playback
	logger   *zerolog.Logger
}

func (l *launcher) Start() {
	l.playback.Start()
	l.profiler.Start()
}

func (l *launcher) Stop() {
	l.logger.Debug().Msg("Stopping playback")
	l.playback.Stop()
	l.logger.Debug().Msg("Stopping profiler")
	l.profiler.Stop()
}

// New runs new service that is used only for integration tests
func New(logger *zerolog.Logger, cfg *config.Config, errChan chan<- error) (common.Service, error) {

	// create memprofiler profiler
	profilerLogger := client.LoggerFromZeroLog(logger)
	profiler, err := client.NewProfiler(profilerLogger, cfg.Client)
	if err != nil {
		return nil, err
	}

	// run memory consumption scenario
	pb := playback.New(logger, cfg.Scenario, errChan)

	l := &launcher{
		profiler: profiler,
		playback: pb,
		logger:   logger,
	}

	return l, nil
}
