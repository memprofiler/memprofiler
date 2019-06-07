package config

import "fmt"

// BackendConfig contains settings for GRPC service - an entry point for incoming data streams
type BackendConfig struct {
	ListenEndpoint string `yaml:"listen_endpoint"`
}

// Verify checks config
func (c *BackendConfig) Verify() error {
	if c.ListenEndpoint == "" {
		return fmt.Errorf("empty BackendConfig.ListenEndpoint")
	}
	return validateEndpoint(c.ListenEndpoint)
}
