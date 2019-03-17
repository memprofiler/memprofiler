package lib

import (
	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/server/common"
)

type launcher struct {
	client   client.Profiler
	playback playback
}

func (l *launcher) Quit() {
	l.playback.Quit()
	l.client.Quit()
}

func NewLauncher(cfg *Config, logger logrus.FieldLogger, errChan chan<- error) (common.Subsystem, error) {

	// create memprofiler client
	profilerLogger := client.LoggerFromLogrus(logger)
	memprofilerClient, err := client.NewProfiler(profilerLogger, cfg.Client)
	if err != nil {
		return nil, err
	}

	// run memory consumption scenario
	playback := newPlayback(newContainer(), cfg.Scenario, errChan)

	l := &launcher{
		client:   memprofilerClient,
		playback: playback,
	}

	return l, nil
}
