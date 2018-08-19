package config

import "fmt"

// WebConfig contains settings for server providing WebUI
type WebConfig struct {
	ListenEndpoint string `yaml:"listen_endpoint"`
}

// Verify checks config
func (c *WebConfig) Verify() error {
	if c.ListenEndpoint == "" {
		return fmt.Errorf("empty WebConfig.ListenEndpoint")
	}

	return validateEndpoint(c.ListenEndpoint)
}
