package config

import (
	"fmt"
	"sort"
	"time"
)

// MetricsConfig contains settings for a task runner
// which computes session metrics in a background
type MetricsConfig struct {
	// AveragingWindows defines which parts of time series will
	// be used for session metrics computation. Client may want to have
	// trend values for last 5 sec, 1 min and 1 hour, for example.
	AveragingWindows []time.Duration `yaml:"averaging_windows"`
}

// Verify checks config
func (c *MetricsConfig) Verify() error {

	if len(c.AveragingWindows) < 1 {
		return fmt.Errorf("no averaging_windows configured")
	}

	if len(c.AveragingWindows) > 5 {
		return fmt.Errorf("too many (more than 5) averaging_windows configured: this will cause high CPU consumption")
	}

	sort.Slice(c.AveragingWindows, func(i, j int) bool { return c.AveragingWindows[i] < c.AveragingWindows[j] })

	return nil
}
