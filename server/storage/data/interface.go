package data

import (
	"context"
	"io"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
)

// Storage provides interface for storing and loading service measurements
// FIXME: pass context with injected logger
type Storage interface {
	NewDataSaver(description *schema.InstanceDescription) (Saver, error)
	NewDataLoader(*schema.SessionDescription) (Loader, error)
	common.Subsystem
}

// Saver is responsible for saving service instance data into the storage
type Saver interface {
	// Save puts measurement into persistent storage
	Save(*schema.Measurement) error
	// Session returns session assigned to the saver
	SessionDescription() *schema.SessionDescription
	io.Closer
}

// Loader provides stream with measurements for a particular session
type Loader interface {
	// Load loads all the measurements that belong to the particular service;
	Load(context.Context) (<-chan *LoadResult, error)
	io.Closer
}

// LoadResult is a sum type for a result of a measurement load operation
type LoadResult struct {
	Measurement *schema.Measurement
	Err         error
}
