package api

import (
	"github.com/vitalyisaev2/memprofiler/schema"
)

var _ saveState = (*saveStateAwaitDescription)(nil)

type saveStateAwaitDescription struct {
	saveStateCommon
}

func (s *saveStateAwaitDescription) addDescription(desc *schema.ServiceDescription) error {
	if err := s.p.setDescription(desc); err != nil {
		s.switchState(finished)
		return err
	}

	dataSaver, err := s.p.getStorage().NewDataSaver(desc)
	if err != nil {
		return err
	}

	newState := &saveStateAwaitMeasurement{
		saveStateCommon: saveStateCommon{code: awaitMeasurement, p: s.p},
		dataSaver:       dataSaver,
	}
	s.p.setState(newState)
	return nil
}
