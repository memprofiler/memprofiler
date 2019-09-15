package client

import (
	"fmt"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

// Config holds various settings for memprofiler client
type Config struct {
	// Remote memprofiler server address
	ServerEndpoint string `json:"server_endpoint" yaml:"server_endpoint"`
	// ServiceDescription will be used to identify data on the server side
	InstanceDescription *schema.InstanceDescription `json:"instance_description" yaml:"instance_description"`
	// Periodicity sets time interval between measurements
	Periodicity *utils.Duration `json:"periodicity" yaml:"periodicity"`
	// Verbose enables measurement printing into logger
	Verbose bool `json:"verbose" yaml:"verbose"`
}

// Verify checks the config
func (c *Config) Verify() error {
	if c.InstanceDescription == nil {
		return fmt.Errorf("empty instance_description")
	}
	if c.InstanceDescription.ServiceName == "" {
		return fmt.Errorf("empty instance_description.service_name")
	}
	if c.InstanceDescription.InstanceName == "" {
		return fmt.Errorf("empty instance_description.instance_name")
	}
	return nil
}
