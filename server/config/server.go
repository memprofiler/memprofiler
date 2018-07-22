package config

import "fmt"

type ServerConfig struct {
	ListenEnpdoint string `yaml:"listen_endpoint"`
}

func (c *ServerConfig) Verify() error {
	if c.ListenEnpdoint == "" {
		return fmt.Errorf("empty ServerConfig.ListenEnpdoint")
	}

	return validateEndpoint(c.ListenEnpdoint)
}
