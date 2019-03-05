package metrics

import (
	"context"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/memprofiler/memprofiler/schema"
)

var (
	stubLogger *logrus.Logger
	stubCtx    = context.Background()
	errOK      error
)

func init() {
	stubLogger = logrus.New()
	stubLogger.Out = ioutil.Discard
}

// A trivial case, when every indicator is incremented once per second within a single location
func TestSessionData_LinearGrowth(t *testing.T) {

	cs := &schema.Callstack{
		Frames: []*schema.StackFrame{
			{File: "a.go", Line: 1},
			{File: "b.go", Line: 2},
		},
		Id: "abcd",
	}

	// make 4 timestamps
	var tstamps []*timestamp.Timestamp
	start := time.Now().Add(-1 * 30 * time.Second)
	step := 10 * time.Second
	for i := 0; i < 4; i++ {
		tstamp, err := ptypes.TimestampProto(start.Add(time.Duration(i) * step))
		if err != nil {
			assert.FailNow(t, "Can not construct timestamp: %v", err)
		}
		tstamps = append(tstamps, tstamp)
	}

	// declare measurements
	mms := []*schema.Measurement{
		{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 0, AllocObjects: 0, FreeBytes: 0, FreeObjects: 0},
					Callstack:   cs,
				},
			},
			ObservedAt: tstamps[0],
		},
		{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1, AllocObjects: 1, FreeBytes: 1, FreeObjects: 1},
					Callstack:   cs,
				},
			},
			ObservedAt: tstamps[1],
		},
		{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 20, AllocObjects: 20, FreeBytes: 20, FreeObjects: 20},
					Callstack:   cs,
				},
			},
			ObservedAt: tstamps[2],
		},
		{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 30, AllocObjects: 30, FreeBytes: 30, FreeObjects: 30},
					Callstack:   cs,
				},
			},
			ObservedAt: tstamps[3],
		},
	}

	// going to estimate heap consumption rates for last 5, 20 and 60 seconds
	fiveSeconds := 5 * time.Second
	twentySeconds := 20 * time.Second
	sixtySeconds := time.Minute
	averagingWindows := []time.Duration{fiveSeconds, twentySeconds, sixtySeconds}

	container := newSessionData(stubLogger, averagingWindows)
	for _, mm := range mms {
		err := container.registerMeasurement(mm)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	// obtain result
	sessionMetrics := container.getSessionMetrics()
	assert.Len(t, sessionMetrics.Locations, 1)
	lm := sessionMetrics.Locations[0]
	assert.Equal(t, cs, lm.Callstack)
	if !assert.NotNil(t, lm.Rates) {
		assert.FailNow(t, "empty location metrics rates")
	}

	// no measurements fall into the last 5 seconds interval, so all trends couldn't be counted
	assert.Equal(t, lm.Rates[0].Span, ptypes.DurationProto(fiveSeconds))
	fiveSecondRates := lm.Rates[0].Values
	assert.True(t, math.IsNaN(fiveSecondRates.AllocBytes))
	assert.True(t, math.IsNaN(fiveSecondRates.AllocObjects))
	assert.True(t, math.IsNaN(fiveSecondRates.FreeBytes))
	assert.True(t, math.IsNaN(fiveSecondRates.FreeObjects))
	assert.True(t, math.IsNaN(fiveSecondRates.InUseBytes))
	assert.True(t, math.IsNaN(fiveSecondRates.InUseObjects))

	// estimate trends for last 20 seconds, it should be about 1 byte per second or 1 unit per second
	assert.Equal(t, lm.Rates[1].Span, ptypes.DurationProto(twentySeconds))
	twentySecondRates := lm.Rates[1].Values
	assert.Equal(t, float64(1), twentySecondRates.AllocBytes)
	assert.Equal(t, float64(1), twentySecondRates.AllocObjects)
	assert.Equal(t, float64(1), twentySecondRates.FreeBytes)
	assert.Equal(t, float64(1), twentySecondRates.FreeObjects)
	assert.Equal(t, float64(0), twentySecondRates.InUseBytes) // mutually compensated
	assert.Equal(t, float64(0), twentySecondRates.InUseObjects)

	// estimate trends for last 60 seconds
	assert.Equal(t, lm.Rates[2].Span, ptypes.DurationProto(sixtySeconds))
	sixtySecondRates := lm.Rates[2].Values
	assert.Equal(t, float64(1.09), sixtySecondRates.AllocBytes)
	assert.Equal(t, float64(1.09), sixtySecondRates.AllocObjects)
	assert.Equal(t, float64(1.09), sixtySecondRates.FreeBytes)
	assert.Equal(t, float64(1.09), sixtySecondRates.FreeObjects)
	assert.Equal(t, float64(0), sixtySecondRates.InUseBytes) // mutually compensated
	assert.Equal(t, float64(0), sixtySecondRates.InUseObjects)
}
