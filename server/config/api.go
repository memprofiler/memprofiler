package config

import "fmt"

type APIConfig struct {
	ListenEnpdoint string `yaml:"listen_endpoint"`
}

func (c *APIConfig) Verify() error {
	if c.ListenEnpdoint == "" {
		return fmt.Errorf("empty APIConfig.ListenEnpdoint")
	}

	return validateEndpoint(c.ListenEnpdoint)
}
