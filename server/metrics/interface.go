package metrics

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// Runner executes statistical analysis of the incoming data stream
type Runner interface {
	PutMeasurement(sd *storage.SessionDescription, mm *schema.Measurement) error
	GetSessionMetrics(ctx context.Context, sd *storage.SessionDescription) (*schema.SessionMetrics, error)
	common.Subsystem
}
