package memprofiler

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	profileRecords = 256
)

type Logger interface {
	Debug(string)
	Error(string)
}

type Profiler interface {
	Quit()
}

type defaultProfiler struct {
	limiter *rate.Limiter

	cfg    *Config
	logger Logger
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *defaultProfiler) loop() {
	defer p.wg.Done()

	for {
		if err := p.limiter.Wait(p.ctx); err != nil {
			return
		}

		mm, err := p.measure()
		if err != nil {
			msg := fmt.Sprintf("Failed to obtain memory profile: %s", err.Error())
			p.logger.Error(msg)
		}

		if p.cfg.DumpToLogger {
			p.dumpMeasurement(mm)
		}
	}
}

func (p *defaultProfiler) measure() (*Measurement, error) {

	stacks := make(map[string]*Location)
	records := getMemProfileRecords()

	for _, record := range records {
		s := &Stack{}
		s.fill(record.Stack(), false)
		id, err := s.hash()
		if err != nil {
			return nil, err
		}

		location, exists := stacks[id]
		if !exists {
			location = &Location{Stack: s, MemoryUsage: &MemoryUsage{}}
			stacks[id] = location
		}

		location.MemoryUsage.update(&record)
	}

	mm := &Measurement{
		Timestamp: time.Now().Unix(),
		Locations: make([]*Location, len(stacks)),
	}

	for _, location := range stacks {
		mm.Locations = append(mm.Locations, location)
	}

	return mm, nil
}

func (p *defaultProfiler) dumpMeasurement(mm *Measurement) {

}

func (p *defaultProfiler) Quit() {
	p.cancel()
	p.wg.Wait()
}

func getMemProfileRecords() []runtime.MemProfileRecord {

	// we don't know how much should we allocate in order to keep profiler dump
	rs := make([]runtime.MemProfileRecord, profileRecords)
	n, ok := runtime.MemProfile(rs, true)
	if !ok {
		for {
			rs = make([]runtime.MemProfileRecord, n+profileRecords)
			n, ok = runtime.MemProfile(rs, true)
			if ok {
				rs = rs[0:n]
				break
			}
		}
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].InUseBytes() > rs[j].InUseBytes() })

	return rs
}

// NewProfiler launches new instance of memory profiler
func NewProfiler(logger Logger, cfg *Config) Profiler {
	ctx, cancel := context.WithCancel(context.Background())
	p := &defaultProfiler{
		limiter: rate.NewLimiter(rate.Every(cfg.Periodicity), 1),
		logger:  logger,
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
	}
	p.wg.Add(1)
	go p.loop()
	return p
}
