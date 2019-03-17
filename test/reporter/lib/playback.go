package lib

import (
	"context"
	"sync"
	"time"

	"github.com/memprofiler/memprofiler/server/common"
)

// playback is responsible for reproducing the desired memory
// consumption behaviour (according to provided scenario)
type playback interface {
	common.Subsystem
}

type defaultPlayback struct {
	container container
	scenario  *Scenario
	errChan   chan<- error // report fatal errors to the client
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func (p *defaultPlayback) Quit() {
	p.cancel()
	p.wg.Wait()
}

func (p *defaultPlayback) loop() {

	defer p.wg.Done()

	for i := 0; ; i++ {

		// pick next available step
		index := i % len(p.scenario.Steps)
		step := p.scenario.Steps[index]

		// change amount of occupied memory and sleep
		if err := p.container.grow(step.MemoryDelta); err != nil {
			select {
			case <-p.ctx.Done():
			case p.errChan <- err:
			}
			return
		}

		// wait for a while
		select {
		case <-time.NewTimer(step.Wait.Duration).C:
			break
		case <-p.ctx.Done():
			return
		}
	}
}

func newPlayback(container container, scenario *Scenario, errChan chan<- error) playback {
	ctx, cancel := context.WithCancel(context.Background())

	pb := &defaultPlayback{
		container: container,
		scenario:  scenario,
		errChan:   errChan,
		wg:        sync.WaitGroup{},
		ctx:       ctx,
		cancel:    cancel,
	}

	pb.wg.Add(1)
	go pb.loop()

	return pb
}
