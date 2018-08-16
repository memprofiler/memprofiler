package api

import (
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ saveState = (*saveStateAwaitMeasurement)(nil)

type saveStateAwaitMeasurement struct {
	saveStateCommon
	dataSaver storage.DataSaver
	counter   int
}

func (s *saveStateAwaitMeasurement) addMeasurement(mm *schema.Measurement) error {
	s.counter++
	s.p.getLogger().WithField("id", s.counter).Debug("Measurement received")
	return s.dataSaver.Save(mm)
}

func (s *saveStateAwaitMeasurement) close() error {
	return s.dataSaver.Close()
}
