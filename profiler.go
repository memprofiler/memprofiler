package analyzer

import (
	"runtime"
	"sort"
	"time"

	"golang.org/x/time/rate"
)

const (
	profileRecords = 256
)

type Profiler interface {
}

type defaultProfiler struct {
	limiter *rate.Limiter
	counter int
}

func (p *defaultProfiler) measure() {

	// 
	stacks := make(map[string]*Location)
	records := getMemProfileRecords()
	timestamp := time.Now()

	for i, record := range p {
		s := &Stack{}
		s.fill(record.Stack(), false)
		id := s.hash()

		location, exists := stacks[id]; 
		!exists {
			location = &Location{Stack: s}
			
		}
	}
}

func getMemProfileRecords() []runtime.MemProfileRecord {

	// we don't know how much should we allocate in order to keep profiler dump
	rs := make([]runtime.MemProfileRecord, profileRecords)
	n, ok := runtime.MemProfile(p, true)
	if !ok {
		for {
			rs = make([]runtime.MemProfileRecord, n+profileRecords)
			n, ok = runtime.MemProfile(p, true)
			if ok {
				rs = rs[0:n]
				break
			}
		}
	}
	sort.Slice(rs, func(i, j int) bool { return p[i].InUseBytes() > p[j].InUseBytes() })

	return rs
}
