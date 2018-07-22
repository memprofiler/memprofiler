package storage

import (
	"context"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
)

// Service provides interface for storing and loading discrete measurements
type Service interface {

	// Save stores measurement in a persistant storage
	SaveMeasurement(*schema.ServiceDescription, *schema.Measurement) error

	// LoadAll loads all the measurements that belong to the particular service;
	// since it's async task, it may be canceled via context
	LoadAllMeasurements(context.Context, *schema.ServiceDescription) (<-chan *LoadResult, error)

	// LoadServices provides list of services available in storage
	LoadServices() ([]*schema.ServiceDescription, error)

	common.Subsystem
}

// LoadResult is a sum type for a result of a measurement load operation
type LoadResult struct {
	Measurement *schema.Measurement
	Err         error
}
