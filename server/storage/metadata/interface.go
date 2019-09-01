package metadata

import (
	"context"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
)

// Storage stores metadata about service measurements
type Storage interface {
	GetServices(ctx context.Context) ([]string, error)
	GetInstances(ctx context.Context, service string) ([]*schema.InstanceDescription, error)
	GetSessions(ctx context.Context, description *schema.InstanceDescription) ([]*schema.Session, error)
	StartSession(ctx context.Context, description *schema.InstanceDescription) (*schema.SessionDescription, error)
	StopSession(ctx context.Context, description *schema.SessionDescription) error
	common.Subsystem
}
