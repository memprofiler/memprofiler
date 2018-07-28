package storage

import (
	"context"
	"io"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
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
	NewDataLoader(*schema.ServiceDescription, SessionID) (DataLoader, error)
}

// MetadataStorage keeps metainformation about service measurements
type MetadataStorage interface {
	Services() []string
	Instances(string) []string
	Sessions(*schema.ServiceDescription) []SessionID
}

// DataSaver is responsible for saving service instance data into the storage
type DataSaver interface {
	// Save puts measurement into persistant storage
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

// ServiceMeta provides metainformation about stored service data
type ServiceMeta struct {
	Description *schema.ServiceDescription
	Sessions    []SessionID
}
