package backend

//go:generate stringer -type=saveStateCode ./save_protocol.go

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/locator"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage/data"
)

// saveState implements state pattern for handling save requests
type saveState interface {
	addDescription(description *schema.InstanceDescription) error
	addMeasurement(mm *schema.Measurement) error
	close() error
}

// saveProtocol provides interface for handling save requests
type saveProtocol interface {
	saveState
	setState(saveState)
	setSessionDescription(*schema.SessionDescription) error
	getSessionDescription() *schema.SessionDescription
	getStorage() data.Storage
	getComputer() metrics.Computer
	setLogger(logger *zerolog.Logger)
	getLogger() *zerolog.Logger
	setDataSaver(data.Saver) error
	getDataSaver() data.Saver
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
	sessionDescription *schema.SessionDescription
	dataSaver          data.Saver
	storage            data.Storage
	computer           metrics.Computer
	logger             *zerolog.Logger
}

var _ saveProtocol = (*defaultSaveProtocol)(nil)

func (p *defaultSaveProtocol) setState(instance saveState) {
	p.saveState = instance
}

func (p *defaultSaveProtocol) setSessionDescription(desc *schema.SessionDescription) error {
	if p.sessionDescription != nil {
		return fmt.Errorf("session description is already set")
	}
	p.sessionDescription = desc
	return nil
}

func (p *defaultSaveProtocol) getSessionDescription() *schema.SessionDescription {
	return p.sessionDescription
}

func (p *defaultSaveProtocol) getComputer() metrics.Computer { return p.computer }

func (p *defaultSaveProtocol) getStorage() data.Storage { return p.storage }

func (p *defaultSaveProtocol) getLogger() *zerolog.Logger { return p.logger }

func (p *defaultSaveProtocol) setLogger(l *zerolog.Logger) { p.logger = l }

func (p *defaultSaveProtocol) setDataSaver(dataSaver data.Saver) error {
	if p.dataSaver != nil {
		return fmt.Errorf("data saver is already set")
	}
	p.dataSaver = dataSaver
	return nil
}

func (p *defaultSaveProtocol) getDataSaver() data.Saver { return p.dataSaver }

func newSaveProtocol(locator *locator.Locator) saveProtocol {

	p := &defaultSaveProtocol{
		storage:  locator.DataStorage,
		computer: locator.Computer,
		logger:   locator.Logger,
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
