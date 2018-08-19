package config

import "fmt"

// MetricsConfig contains settings for a task runner
// which computes session metrics in a background
type MetricsConfig struct {
	// Window defines which part of time series will
	// be used for session metrics computation
	Window int `yaml:"window"`
}

// Verify checks config
func (c *MetricsConfig) Verify() error {
	if c.Window <= 0 {
		return fmt.Errorf("invalid MetricsConfig.Window: %d", c.Window)
	}
	return nil
}
