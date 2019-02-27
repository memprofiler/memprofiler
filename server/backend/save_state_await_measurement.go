package backend

import (
	"github.com/memprofiler/memprofiler/schema"
)

var _ saveState = (*saveStateAwaitMeasurement)(nil)

type saveStateAwaitMeasurement struct {
	saveStateCommon
	counter int
}

func (s *saveStateAwaitMeasurement) addMeasurement(mm *schema.Measurement) error {
	s.counter++
	s.p.getLogger().WithField("id", s.counter).Debug("Measurement received")

	// 1. Save data to persistant storage
	if err := s.p.getDataSaver().Save(mm); err != nil {
		return err
	}

	// 2. Save measurement to metrics computer
	return s.p.getComputer().PutMeasurement(s.p.getSessionDescription(), mm)
}

func (s *saveStateAwaitMeasurement) close() error {
	return s.p.getDataSaver().Close()
}
