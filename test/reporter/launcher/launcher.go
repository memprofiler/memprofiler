package launcher

import (
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/test/reporter/config"
	"github.com/memprofiler/memprofiler/test/reporter/playback"
)

type launcher struct {
	profiler client.Profiler
	playback playback.Playback
	logger   logrus.FieldLogger
}

func (l *launcher) Start() {
	l.playback.Start()
	l.profiler.Start()
}

func (l *launcher) Stop() {
	l.logger.Debug("Stopping playback")
	l.playback.Stop()
	l.logger.Debug("Stopping profiler")
	l.profiler.Stop()
}

func New(logger logrus.FieldLogger, cfg *config.Config, errChan chan<- error) (common.Service, error) {

	// create memprofiler profiler
	profilerLogger := client.LoggerFromLogrus(logger)
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
