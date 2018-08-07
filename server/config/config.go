package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config is a top-level structure with all server settings
type Config struct {
	Server  *ServerConfig  `yaml:"server"`
	Storage *StorageConfig `yaml:"storage"`
	Logging *LoggingConfig `yaml:"logging"`
	Metrics *MetricsConfig `yaml:"metrics"`
}

func (c *Config) Verify() error {
	// TODO: use reflect to iterate over pointers
	if err := c.Server.Verify(); err != nil {
		return err
	}
	if err := c.Storage.Verify(); err != nil {
		return err
	}
	if err := c.Logging.Verify(); err != nil {
		return err
	}
	return nil
}

// NewConfigFromFile builds config structure from path
func NewConfigFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err = yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	if err = c.Verify(); err != nil {
		return nil, err
	}

	return &c, nil
}
