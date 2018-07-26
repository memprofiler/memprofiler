package api

//go:generate stringer -type=saveStateCode ./save_protocol.go

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/locator"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

// saveState implements state pattern for handling save requests
type saveState interface {
	addDescription(*schema.ServiceDescription) error
	addMeasurement(*schema.Measurement) error
}

// saveProtocol provides interface for handling save requests
type saveProtocol interface {
	saveState
	setState(saveState)
	setDescription(*schema.ServiceDescription) error
	getDescription() *schema.ServiceDescription
	getStorage() storage.Storage
}

type saveStateCode int8

const (
	awaitHeader saveStateCode = iota + 1
	awaitMeasurement
	broken
)

// defaultSaveProtocol is a default implementation of saveProtocol
type defaultSaveProtocol struct {
	saveState
	desc    *schema.ServiceDescription
	storage storage.Storage
	logger  *logrus.Logger
}

var _ saveProtocol = (*defaultSaveProtocol)(nil)

func (p *defaultSaveProtocol) setState(instance saveState) {
	p.saveState = instance
}

func (p *defaultSaveProtocol) setDescription(desc *schema.ServiceDescription) error {
	if p.desc != nil {
		return fmt.Errorf("description is already set")
	}
	p.desc = desc
	return nil
}

func (p *defaultSaveProtocol) getDescription() *schema.ServiceDescription {
	return p.desc
}

func (p *defaultSaveProtocol) getStorage() storage.Storage {
	return p.storage
}

func newSaveProtocol(locator *locator.Locator) saveProtocol {
	p := &defaultSaveProtocol{
		storage: locator.Storage,
		logger:  locator.Logger,
	}
	return p
}
