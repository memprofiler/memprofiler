package client

import (
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

// Config holds various settings for memprofiler client
type Config struct {
	// Remote memprofiler server address
	ServerEndpoint string `json:"server_endpoint" yaml:"server_endpoint"`
	// ServiceDescription will be used to identify data on the server side
	ServiceDescription *schema.ServiceDescription `json:"service_description" yaml:"service_description"`
	// Periodicity sets time interval between measurements
	Periodicity utils.Duration `json:"duration" yaml:"periodicity"`
	// Verbose enables measurement printing into logger
	Verbose bool `json:"verbose" yaml:"verbose"`
}
