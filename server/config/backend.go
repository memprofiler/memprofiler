package config

import "fmt"

// BackendConfig contains settings for GRPC service - an entry point for incoming data streams
type BackendConfig struct {
	ListenEnpdoint string `yaml:"listen_endpoint"`
}

// Verify checks config
func (c *BackendConfig) Verify() error {
	if c.ListenEnpdoint == "" {
		return fmt.Errorf("empty BackendConfig.ListenEnpdoint")
	}
	return validateEndpoint(c.ListenEnpdoint)
}
