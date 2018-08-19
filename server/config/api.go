package config

import "fmt"

// APIConfig contains settings for GRPC service - an entry point for incoming data streams
type APIConfig struct {
	ListenEnpdoint string `yaml:"listen_endpoint"`
}

// Verify verifies config
func (c *APIConfig) Verify() error {
	if c.ListenEnpdoint == "" {
		return fmt.Errorf("empty APIConfig.ListenEnpdoint")
	}
	return validateEndpoint(c.ListenEnpdoint)
}
