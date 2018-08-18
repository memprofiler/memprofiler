package metrics

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
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
func TestDefaultComputer_ComputeSessionMetrics_LinearGrowth(t *testing.T) {
	c := New(stubLogger)

	cs := &schema.CallStack{Frames: []*schema.StackFrame{{File: "a", Line: 1}}}
	ch := make(chan *storage.LoadResult, 3)
	ch <- &storage.LoadResult{
		Measurement: &schema.Measurement{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 0, AllocObjects: 0, FreeBytes: 0, FreeObjects: 0},
					CallStack:   cs,
				},
			},
			ObservedAt: &timestamp.Timestamp{Seconds: 0},
		},
	}
	ch <- &storage.LoadResult{
		Measurement: &schema.Measurement{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1, AllocObjects: 1, FreeBytes: 1, FreeObjects: 1},
					CallStack:   cs,
				},
			},
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
		},
	}
	ch <- &storage.LoadResult{
		Measurement: &schema.Measurement{
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2, AllocObjects: 2, FreeBytes: 2, FreeObjects: 2},
					CallStack:   cs,
				},
			},
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
		},
	}
	close(ch)

	dl := &storage.DataLoaderMock{}
	dl.On("Load", stubCtx).Return(ch, errOK).Once()
	dl.On("Close").Return(errOK).Once()

	sm, err := c.SessionMetrics(stubCtx, dl)
	locations := sm.Locations
	assert.NoError(t, err)
	assert.Len(t, locations, 1)
	assert.Equal(t, cs, locations[0].CallStack)

	// for alloc/free rate is 1 unit per second
	assert.Equal(t, float64(1), locations[0].Average.AllocBytesRate)
	assert.Equal(t, float64(1), locations[0].Average.AllocObjectsRate)
	assert.Equal(t, float64(1), locations[0].Average.FreeBytesRate)
	assert.Equal(t, float64(1), locations[0].Average.FreeObjectsRate)
	// since alloc/free rates are equal, in use rates should be zero
	assert.Equal(t, float64(0), locations[0].Average.InUseBytesRate)
	assert.Equal(t, float64(0), locations[0].Average.InUseObjectsRate)
}
