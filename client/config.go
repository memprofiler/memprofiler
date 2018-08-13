package client

import (
	"time"

	"github.com/vitalyisaev2/memprofiler/schema"
)

// Config holds various settings for memprofiler client
type Config struct {
	// Remote memprofiler server address
	ServerEndpoint string
	// ServiceDescription will be used to identify data on the server side
	ServiceDescription *schema.ServiceDescription
	// Periodicity sets time interval between measurements
	Periodicity time.Duration
	// Verbose enables measurement printing into logger
	Verbose bool
}
