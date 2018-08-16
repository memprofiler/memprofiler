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
	close() error
}

// saveProtocol provides interface for handling save requests
type saveProtocol interface {
	saveState
	setState(saveState)
	setDescription(*schema.ServiceDescription) error
	getDescription() *schema.ServiceDescription
	getStorage() storage.Storage
	getLogger() logrus.FieldLogger
	setLogger(logrus.FieldLogger)
}

type saveStateCode int8

const (
	awaitDescription saveStateCode = iota + 1
	awaitMeasurement
	finished
)

// defaultSaveProtocol is a default implementation of saveProtocol
type defaultSaveProtocol struct {
	saveState
	desc    *schema.ServiceDescription
	storage storage.Storage
	logger  logrus.FieldLogger
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

func (p *defaultSaveProtocol) getLogger() logrus.FieldLogger { return p.logger }

func (p *defaultSaveProtocol) setLogger(l logrus.FieldLogger) { p.logger = l }

func newSaveProtocol(locator *locator.Locator) saveProtocol {

	p := &defaultSaveProtocol{
		storage: locator.Storage,
		logger:  locator.Logger,
	}

	// waiting for header message first
	p.saveState = &saveStateAwaitDescription{
		saveStateCommon: saveStateCommon{
			p:    p,
			code: awaitDescription,
		},
	}
	return p
}
