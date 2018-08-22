package metrics

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// Computer performs statistical analysis for the incoming and archived data streams
type Computer interface {
	PutMeasurement(sd *storage.SessionDescription, mm *schema.Measurement) error
	GetSessionMetrics(ctx context.Context, sd *storage.SessionDescription) (*schema.SessionMetrics, error)
	common.Subsystem
}
