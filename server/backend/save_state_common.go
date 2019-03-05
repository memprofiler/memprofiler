package backend

import (
	"fmt"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

var _ saveState = (*saveStateCommon)(nil)

type saveStateCommon struct {
	code saveStateCode
	p    saveProtocol
}

func (s *saveStateCommon) addDescription(*schema.ServiceDescription) error {
	return s.makeError()
}

func (s *saveStateCommon) addMeasurement(*schema.Measurement) error {
	return s.makeError()
}

func (s *saveStateCommon) close() error {
	if dataSaver := s.p.getDataSaver(); dataSaver != nil {
		return dataSaver.Close()
	}
	return nil
}

func (s *saveStateCommon) makeError() error {
	defer s.switchState(finished)
	return fmt.Errorf(
		"unexpected call of method %s for state %s",
		utils.Caller(2), s.code.String(),
	)
}

func (s *saveStateCommon) switchState(code saveStateCode) {
	var newState saveState
	switch code {
	case awaitDescription:
		newState = &saveStateAwaitDescription{saveStateCommon: saveStateCommon{p: s.p, code: code}}
	case awaitMeasurement:
		newState = &saveStateAwaitMeasurement{saveStateCommon: saveStateCommon{p: s.p, code: code}}
	case finished:
		newState = &saveStateFinished{saveStateCommon: saveStateCommon{p: s.p, code: code}}
	}

	s.p.setState(newState)
}
