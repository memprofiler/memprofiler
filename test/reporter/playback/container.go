package playback

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// container provides interface to alter the amount of consumed memory
type container interface {
	// grow changes the size of the memory consumed;
	// (delta may be negative as well, but overall size can't go beyond zero)
	grow(delta int) error
}

// defaultContainer just holds some object in memory
type defaultContainer struct {
	logger logrus.FieldLogger
	array  []int
}

func (c *defaultContainer) grow(delta int) error {
	target := len(c.array) + delta
	if target < 0 {
		return fmt.Errorf("target value is below zero (curr: %d, delta: %d)", len(c.array), delta)
	}
	c.logger.WithFields(logrus.Fields{"size": len(c.array), "delta": delta}).Debug("Growing memory")
	c.array = make([]int, target)
	return nil
}

// newContainer constructs new container
func newContainer(logger logrus.FieldLogger) container { return &defaultContainer{logger: logger} }
