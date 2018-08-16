package config

import "fmt"

type WebConfig struct {
	ListenEndpoint string `yaml:"listen_endpoint"`
}

func (c *WebConfig) Verify() error {
	if c.ListenEndpoint == "" {
		return fmt.Errorf("empty WebConfig.ListenEndpoint")
	}

	return validateEndpoint(c.ListenEndpoint)
}
