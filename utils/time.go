package utils

import (
	"math"
	"time"
)

const (
	nanosecondsInSecond = float64(time.Second / time.Nanosecond)
)

// TimeToFloat64 converts timestamp to float64 (in seconds)
func TimeToFloat64(tstamp time.Time) float64 {
	return float64(tstamp.UnixNano()) / nanosecondsInSecond
}

// Float64ToTime converts float time (seconds) to *time.Time
// FIXME: unused, remove?
func Float64ToTime(f float64) time.Time {
	sec, nsec := math.Modf(f)
	result := time.Unix(int64(sec), int64(nsec*nanosecondsInSecond))
	return result
}
