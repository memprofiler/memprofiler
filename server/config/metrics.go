package config

import (
	"fmt"
	"sort"
	"time"

	"github.com/ChannelMeter/iso8601duration"
)

// MetricsConfig contains settings for a task runner
// which computes session metrics in a background
type MetricsConfig struct {
	// AveragingWindowsString defines which parts of time series will
	// be used for session metrics computation. Client may want to have
	// trend values for last 5 sec, 1 min and 1 hour, for example.
	AveragingWindowsString []string `yaml:"averaging_windows"`
	// AveragingWindow is a sorted list of averaging window values
	AveragingWindows []time.Duration `yaml:"-"`
}

// Verify checks config
func (c *MetricsConfig) Verify() error {

	if len(c.AveragingWindowsString) < 1 {
		return fmt.Errorf("empty MetricsConfig.AveragingWindow")
	}

	// fill duration list with data
	for _, windowStr := range c.AveragingWindowsString {
		windowDuration, err := duration.FromString(windowStr)
		if err != nil {
			return fmt.Errorf("failed to parse string '%s' to duration: %v", windowStr, err)
		}
		c.AveragingWindows = append(c.AveragingWindows, windowDuration.ToDuration())
	}
	sort.Slice(c.AveragingWindows, func(i, j int) bool {
		return c.AveragingWindows[i] < c.AveragingWindows[j]
	})

	return nil
}
