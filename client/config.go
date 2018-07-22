package memprofiler

import "time"

type Config struct {
	// ProcessID helps to distinguish memory profile produced by this particular process
	// (you can use app name, IP, node_id, whatever)
	ProcessID string
	// Periodicity sets time interval between memory usage measurements
	Periodicity time.Duration
	// DumpToLogger enables measurement printing into logger
	DumpToLogger bool
}
