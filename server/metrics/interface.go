package metrics

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// Computer is responsible for counting memory usage metrics from data
type Computer interface {
	ComputeSessionMetrics(context.Context, storage.DataLoader) ([]*LocationMetrics, error)
	common.Subsystem
}

// HeapConsumptionRates is a collection of rate values for memory consumption indicators.
// For bytes rate units are bytes per second, for Objects units are units per second
type HeapConsumptionRates struct {
	InUseObjectsRate float64
	InUseBytesRate   float64
	FreeObjectsRate  float64
	FreeBytesRate    float64
	AllocObjectsRate float64
	AllocBytesRate   float64
}

// LocationMetrics is a set of metrics computed for a particular location in code
type LocationMetrics struct {
	// Average represents heap consumption rates taken from beginning of the session
	Average *HeapConsumptionRates
	// Recent represents heap consumption rates taken from several last measurements
	Recent *HeapConsumptionRates
	// CallStack describes location in code where the allocation occured
	CallStack *schema.CallStack
}
