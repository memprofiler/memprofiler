package client

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/memprofiler/memprofiler/server/common"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

// Profiler should keep working during whole application lifetime
type Profiler interface {
	common.Service
}

type defaultProfiler struct {
	stream     schema.MemprofilerBackend_SaveReportClient
	limiter    *rate.Limiter
	clientConn *grpc.ClientConn
	cfg        *Config
	logger     Logger
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func (p *defaultProfiler) loop() {
	defer p.wg.Done()

	for {
		if err := p.limiter.Wait(p.ctx); err != nil {
			// time is out
			break
		}

		// create memory report and stream it to server
		if err := p.report(); err != nil {
			p.logger.Error(err.Error())
		}
	}

	// close stream explicitly
	msg, err := p.stream.CloseAndRecv()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to close stream: %v", err))
	} else {
		p.logger.Debug(fmt.Sprintf("Final stream result: %v", msg))
	}

}

// report gets new measurement and sends it to GRPC stream
func (p *defaultProfiler) report() error {

	// obtain new measurement
	mm, err := p.measure()
	if err != nil {
		return fmt.Errorf("failed to obtain memory profile: %v", err)
	}
	p.maybeDumpMessage(mm)

	// send it to server
	msg := &schema.SaveReportRequest{
		Payload: &schema.SaveReportRequest_Measurement{
			Measurement: mm,
		},
	}
	if err := p.stream.Send(msg); err != nil {
		return fmt.Errorf("failed to send message to server: %v", err)
	}

	return nil
}

// take memory profiling data from runtime
func (p *defaultProfiler) measure() (*schema.Measurement, error) {

	var (
		err     error
		stacks  = make(map[string]*schema.Location)
		records = getMemProfileRecords()
	)

	// iterate over profiler records, prepare structures to be sent to the server
	for _, record := range records {
		cs := &schema.Callstack{}
		utils.FillCallstack(cs, record.Stack(), false)
		cs.Id, err = utils.HashCallstack(cs)
		if err != nil {
			return nil, err
		}

		location, exists := stacks[cs.Id]
		if !exists {
			location = &schema.Location{Callstack: cs, MemoryUsage: &schema.MemoryUsage{}}
			stacks[cs.Id] = location
		}

		utils.UpdateMemoryUsage(location.MemoryUsage, record)
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

func (p *defaultProfiler) maybeDumpMessage(mm *schema.Measurement) {
	if p.cfg.Verbose {
		dump, err := json.Marshal(mm)
		if err != nil {
			p.logger.Error(fmt.Sprintf("Failed to marshal measurement: %s", err.Error()))
		} else {
			p.logger.Debug("Measurement sent: " + string(dump))
		}
	}
}

func (p *defaultProfiler) Start() {
	p.wg.Add(1)
	go p.loop()
}

func (p *defaultProfiler) Stop() {
	p.cancel()
	p.wg.Wait()
	if err := p.clientConn.Close(); err != nil {
		p.logger.Error("Failed to close connection: " + err.Error())
	}
}

// NewProfiler launches new instance of memory profiler
func NewProfiler(logger Logger, cfg *Config) (Profiler, error) {

	// prepare GRPC client
	clientConn, err := grpc.Dial(cfg.ServerEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	stream, err := makeStream(clientConn, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &defaultProfiler{
		stream:     stream,
		limiter:    rate.NewLimiter(rate.Every(cfg.Periodicity.Duration), 1),
		logger:     logger,
		clientConn: clientConn,
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
	}

	return p, nil
}

const profileRecords = 256

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

// makeStream initializes GRPC stream
func makeStream(clientConn *grpc.ClientConn, cfg *Config) (schema.MemprofilerBackend_SaveReportClient, error) {

	c := schema.NewMemprofilerBackendClient(clientConn)

	// open client-side streaming
	stream, err := c.SaveReport(context.Background())
	if err != nil {
		return nil, err
	}

	// send greeting message to server
	msg := &schema.SaveReportRequest{
		Payload: &schema.SaveReportRequest_ServiceDescription{
			ServiceDescription: cfg.ServiceDescription,
		},
	}
	if err := stream.Send(msg); err != nil {
		return nil, fmt.Errorf("failed to send greeting message")
	}

	return stream, nil
}
