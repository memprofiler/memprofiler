package metrics

import (
	"context"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/storage"
)

// Computer performs statistical analysis for the incoming and archived data streams
type Computer interface {
	PutMeasurement(sd *storage.SessionDescription, mm *schema.Measurement) error
	GetSessionMetrics(ctx context.Context, sd *storage.SessionDescription) (*schema.SessionMetrics, error)
	common.Subsystem
}
