package storage

import (
	"context"
	"io"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
)

// Storage keeps service's measurements
type Storage interface {
	DataStorage
	MetadataStorage
	common.Subsystem
}

// DataStorage provides interface for storing and loading service measurements
type DataStorage interface {
	NewDataSaver(*schema.ServiceDescription) (DataSaver, error)
	NewDataLoader(*schema.SessionDescription) (DataLoader, error)
}

// MetadataStorage stores metadata about service measurements
type MetadataStorage interface {
	Services() []string
	Instances(string) ([]*schema.ServiceDescription, error)
	Sessions(*schema.ServiceDescription) ([]*schema.Session, error)
}

// DataSaver is responsible for saving service instance data into the storage
type DataSaver interface {
	// Save puts measurement into persistent storage
	Save(*schema.Measurement) error
	// SessionID returns ID assigned to current steaming session
	SessionID() SessionID
	io.Closer
}

// DataLoader is responsible for obtaining service instance data from storage
type DataLoader interface {
	// Load loads all the measurements that belong to the particular service;
	Load(context.Context) (<-chan *LoadResult, error)
	io.Closer
}

// LoadResult is a sum type for a result of a measurement load operation
type LoadResult struct {
	Measurement *schema.Measurement
	Err         error
}
