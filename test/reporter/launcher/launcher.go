package launcher

import (
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/test/reporter/config"
	"github.com/memprofiler/memprofiler/test/reporter/playback"
)

type launcher struct {
	client   client.Profiler
	playback playback.Playback
}

func (l *launcher) Start() {
	l.playback.Start()
	l.client.Start()
}

func (l *launcher) Stop() {
	l.playback.Stop()
	l.client.Stop()
}

func New(logger logrus.FieldLogger, cfg *config.Config, errChan chan<- error) (common.Service, error) {

	// create memprofiler client
	profilerLogger := client.LoggerFromLogrus(logger)
	memprofilerClient, err := client.NewProfiler(profilerLogger, cfg.Client)
	if err != nil {
		return nil, err
	}

	// run memory consumption scenario
	playback := playback.New(logger, cfg.Scenario, errChan)

	l := &launcher{
		client:   memprofilerClient,
		playback: playback,
	}

	return l, nil
}
