package playback

import (
	"fmt"

	"github.com/rs/zerolog"
)

// container provides interface to alter the amount of consumed memory
type container interface {
	// grow changes the size of the memory consumed;
	// (delta may be negative as well, but overall size can't go beyond zero)
	grow(delta int) error
}

// defaultContainer just holds some object in memory
type defaultContainer struct {
	logger *zerolog.Logger
	array  []int
}

func (c *defaultContainer) grow(delta int) error {
	target := len(c.array) + delta
	if target < 0 {
		return fmt.Errorf("target value is below zero (curr: %d, delta: %d)", len(c.array), delta)
	}
	c.logger.Debug().Fields(map[string]interface{}{
		"size":  len(c.array),
		"delta": delta},
	).Msg("Growing memory")

	c.array = make([]int, target)
	return nil
}

// newContainer constructs new container
func newContainer(logger *zerolog.Logger) container { return &defaultContainer{logger: logger} }
