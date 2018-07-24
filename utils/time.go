package utils

import "time"

func TimeToFloat64(tstamp time.Time) float64 {
	return float64(tstamp.UnixNano())
}

func Float64ToTime(f float64) time.Time {
	sec := int64(f) / (int64(time.Second) / int64(time.Nanosecond))
	nsec := int64(f) % int64(time.Second)
	return time.Unix(sec, nsec)
}
