package api

import "github.com/vitalyisaev2/memprofiler/schema"

var _ saveState = (*saveStateAwaitMeasurement)(nil)

type saveStateAwaitMeasurement struct {
	saveStateCommon
}

func (s *saveStateAwaitMeasurement) addMeasurement(mm *schema.Measurement) error {
	return s.p.getStorage().SaveMeasurement(s.p.getDescription(), mm)
}
