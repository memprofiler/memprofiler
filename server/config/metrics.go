package config

import "time"

// MetricsConfig represents settings for a subsystem responsible for metrics estimation
type MetricsConfig struct {
	// RecentBorder is threshold that helps to determine which events
	// metrics computer should consider as recent
	RecentBorder time.Duration
}
