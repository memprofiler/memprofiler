package metrics

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/memprofiler/schema"
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

	mm0 := &schema.Measurement{
		Locations: []*schema.Location{
			{
				MemoryUsage: &schema.MemoryUsage{AllocBytes: 0, AllocObjects: 0, FreeBytes: 0, FreeObjects: 0},
				Callstack:   cs,
			},
		},
		ObservedAt: &timestamp.Timestamp{Seconds: 0},
	}

	mm1 := &schema.Measurement{
		Locations: []*schema.Location{
			{
				MemoryUsage: &schema.MemoryUsage{AllocBytes: 1, AllocObjects: 1, FreeBytes: 1, FreeObjects: 1},
				Callstack:   cs,
			},
		},
		ObservedAt: &timestamp.Timestamp{Seconds: 1},
	}

	mm2 := &schema.Measurement{
		Locations: []*schema.Location{
			{
				MemoryUsage: &schema.MemoryUsage{AllocBytes: 2, AllocObjects: 2, FreeBytes: 2, FreeObjects: 2},
				Callstack:   cs,
			},
		},
		ObservedAt: &timestamp.Timestamp{Seconds: 2},
	}

	data := newSessionData(stubLogger, 10)
	err := data.registerMeasurement(mm0)
	assert.NoError(t, err)
	err = data.registerMeasurement(mm1)
	assert.NoError(t, err)
	err = data.registerMeasurement(mm2)
	assert.NoError(t, err)

	sm := data.getSessionMetrics()
	assert.Len(t, sm.Locations, 1)
	lm := sm.Locations[0]
	assert.Equal(t, cs, lm.Callstack)

	// for alloc/free rate is 1 unit per second
	assert.Equal(t, float64(1), lm.Rates.AllocBytes)
	assert.Equal(t, float64(1), lm.Rates.AllocObjects)
	assert.Equal(t, float64(1), lm.Rates.FreeBytes)
	assert.Equal(t, float64(1), lm.Rates.FreeObjects)
	// since alloc/free rates are equal, in use rates should be zero
	assert.Equal(t, float64(0), lm.Rates.InUseBytes)
	assert.Equal(t, float64(0), lm.Rates.InUseObjects)
}
