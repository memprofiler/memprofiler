package memprofiler

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/golang/protobuf/ptypes"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
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
			dump, err := json.Marshal(mm)
			if err != nil {
				p.logger.Error(fmt.Sprintf("Failed to marshal measurement: %s", err.Error()))
			} else {
				p.logger.Debug(string(dump))
			}
		}
	}
}

func (p *defaultProfiler) measure() (*schema.Measurement, error) {

	stacks := make(map[string]*schema.Location)
	records := getMemProfileRecords()

	for _, record := range records {
		cs := &schema.CallStack{}
		utils.FillCallStack(cs, record.Stack(), false)
		id, err := utils.HashCallStack(cs)
		if err != nil {
			return nil, err
		}

		location, exists := stacks[id]
		if !exists {
			location = &schema.Location{CallStack: cs, MemoryUsage: &schema.MemoryUsage{}}
			stacks[id] = location
		}

		utils.UpdateMemoryUsage(location.MemoryUsage, &record)
	}

	mm := &schema.Measurement{
		ObservedAt: ptypes.TimestampNow(),
		Locations:  make([]*schema.Location, 0, len(stacks)),
	}

	for _, location := range stacks {
		mm.Locations = append(mm.Locations, location)
	}

	return mm, nil
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
