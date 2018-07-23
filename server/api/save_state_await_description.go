package api

import "github.com/vitalyisaev2/memprofiler/schema"

var _ saveState = (*saveStateAwaitDescription)(nil)

type saveStateAwaitDescription struct {
	saveStateCommon
	protocol saveProtocol
}

func (s *saveStateAwaitDescription) addDescription(desc *schema.ServiceDescription) error {
	if err := s.protocol.setDescription(desc); err != nil {
		s.switchState(broken)
		return err
	}

	s.switchState(awaitMeasurement)
	return nil
}
