package config

import "fmt"

// BackendConfig contains settings for GRPC service - an entry point for incoming data streams
type BackendConfig struct {
	ListenEndpoint string `yaml:"listen_endpoint"`
}

// Verify checks config
func (c *BackendConfig) Verify() error {
	if c == nil {
		return fmt.Errorf("empty backend config")
	}
	if c.ListenEndpoint == "" {
		return fmt.Errorf("empty listen_endpoint")
	}
	return validateEndpoint(c.ListenEndpoint)
}
