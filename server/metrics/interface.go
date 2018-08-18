package metrics

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// Computer is responsible for counting memory usage metrics from data
type Computer interface {
	SessionMetrics(context.Context, storage.DataLoader) ([]*schema.LocationMetrics, error)
	common.Subsystem
}
