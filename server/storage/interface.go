package storage

import (
	"context"
	"io"

	"fmt"

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
	NewDataLoader(sd *SessionDescription) (DataLoader, error)
}

// MetadataStorage stores metadata about service measurements
type MetadataStorage interface {
	Services() []string
	Instances(string) []string
	Sessions(*schema.ServiceDescription) []SessionID
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

// SessionDescription helps to identify streaming session of a particular service instance
type SessionDescription struct {
	ServiceDescription *schema.ServiceDescription
	SessionID          SessionID
}

func (sd SessionDescription) String() string {
	return fmt.Sprintf(
		"%s::%s::%d",
		sd.ServiceDescription.Type,
		sd.ServiceDescription.Instance,
		sd.SessionID,
	)
}
