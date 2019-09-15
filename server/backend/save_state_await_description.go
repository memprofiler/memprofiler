package backend

import (
	"github.com/memprofiler/memprofiler/schema"
)

var _ saveState = (*saveStateAwaitDescription)(nil)

type saveStateAwaitDescription struct {
	saveStateCommon
}

func (s *saveStateAwaitDescription) addDescription(instanceDesc *schema.InstanceDescription) error {

	// run new saver in persistent storage
	dataSaver, err := s.p.getStorage().NewDataSaver(instanceDesc)
	if err != nil {
		s.switchState(finished)
		return err
	}

	if err := s.p.setDataSaver(dataSaver); err != nil {
		s.switchState(finished)
		return err
	}

	// set session description
	if err := s.p.setSessionDescription(dataSaver.SessionDescription()); err != nil {
		s.switchState(finished)
		return err
	}

	// annotate logger and save it for further usage
	logger := s.p.getLogger().With().Fields(map[string]interface{}{
		"service":    instanceDesc.ServiceName,
		"instance":   instanceDesc.InstanceName,
		"session_id": dataSaver.SessionDescription().Id,
	}).Logger()
	s.p.setLogger(&logger)
	logger.Info().Msg("Received greeting message from client")

	s.switchState(awaitMeasurement)
	return nil
}
