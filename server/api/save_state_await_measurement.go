package api

import (
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ saveState = (*saveStateAwaitMeasurement)(nil)

type saveStateAwaitMeasurement struct {
	saveStateCommon
	dataSaver storage.DataSaver
}

func (s *saveStateAwaitMeasurement) addMeasurement(mm *schema.Measurement) error {
	return s.dataSaver.Save(mm)
}
