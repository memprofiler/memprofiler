package config

import "fmt"

// FrontendConfig contains settings for server providing WebUI
type FrontendConfig struct {
	ListenEndpoint string `yaml:"listen_endpoint"`
}

// Verify checks config
func (c *FrontendConfig) Verify() error {
	if c.ListenEndpoint == "" {
		return fmt.Errorf("empty FrontendConfig.ListenEndpoint")
	}

	return validateEndpoint(c.ListenEndpoint)
}
