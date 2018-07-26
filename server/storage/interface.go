package storage

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
)

// Storage provides interface for storing and loading discrete measurements
type Storage interface {
	NewDataSaver(*schema.ServiceDescription) (DataSaver, error)
	NewDataLoader(*schema.ServiceDescription) (DataLoader, error)
	ServiceMeta() ([]*ServiceMeta, error)
	common.Subsystem
}

// DataSaver is responsible for saving service instance data into the storage
type DataSaver interface {
	// Save puts measurement into persistant storage
	Save(*schema.Measurement) error
	// Close should be called when the sender is gone;
	// this call interrupts measurement streaming session
	Close()
	// SessionID returns ID assigned to current steaming session
	SessionID() SessionID
}

// DataLoader is responsible for obtaining service instance data from storage
type DataLoader interface {
	// Load loads all the measurements that belong to the particular service;
	Load(context.Context, SessionID) (<-chan *LoadResult, error)
	// Close should be called when the receiver don't to load data anymore;
	// this call interrupts measurement streaming session
	Close()
}

// LoadResult is a sum type for a result of a measurement load operation
type LoadResult struct {
	Measurement *schema.Measurement
	Err         error
}

// SessionID is a unique identifier for a measurement streaming session;
// all the sessions will be ordered by this string value
type SessionID string

// ServiceMeta provides metainformation about stored service data
type ServiceMeta struct {
	Description *schema.ServiceDescription
	Sessions    []SessionID
}
