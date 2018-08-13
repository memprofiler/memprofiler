package api

import (
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
)

var _ saveState = (*saveStateAwaitDescription)(nil)

type saveStateAwaitDescription struct {
	saveStateCommon
}

func (s *saveStateAwaitDescription) addDescription(desc *schema.ServiceDescription) error {
	// annotate logger
	logger := s.p.getLogger().WithFields(logrus.Fields{
		"type":     desc.GetType(),
		"instance": desc.GetInstance(),
	})
	logger.Info("Received greeting message from client")

	// try to set description
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
