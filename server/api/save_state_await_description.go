package api

import (
	"github.com/sirupsen/logrus"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var _ saveState = (*saveStateAwaitDescription)(nil)

type saveStateAwaitDescription struct {
	saveStateCommon
}

func (s *saveStateAwaitDescription) addDescription(desc *schema.ServiceDescription) error {

	// run new saver in persistent storage
	dataSaver, err := s.p.getStorage().NewDataSaver(desc)
	if err != nil {
		s.switchState(finished)
		return err
	}

	if err := s.p.setDataSaver(dataSaver); err != nil {
		s.switchState(finished)
		return err
	}

	// set session description
	sd := &storage.SessionDescription{
		ServiceDescription: desc,
		SessionID:          dataSaver.SessionID(),
	}
	if err := s.p.setSessionDescription(sd); err != nil {
		s.switchState(finished)
		return err
	}

	// annotate logger and save it for further usage
	logger := s.p.getLogger().WithFields(logrus.Fields{
		"type":       desc.GetType(),
		"instance":   desc.GetInstance(),
		"session_id": dataSaver.SessionID(),
	})
	s.p.setLogger(logger)
	logger.Info("Received greeting message from client")

	s.switchState(awaitMeasurement)
	return nil
}
