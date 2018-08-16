package api

import (
	"fmt"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/utils"
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
	return s.makeError()
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
