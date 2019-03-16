package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config is a top-level structure with all server settings
type Config struct {
	API     *APIConfig     `yaml:"api"`
	Web     *WebConfig     `yaml:"web"`
	Storage *StorageConfig `yaml:"storage"`
	Metrics *MetricsConfig `yaml:"metrics"`
	Logging *LoggingConfig `yaml:"logging"`
}

// Verify checks config
func (c *Config) Verify() error {
	// TODO: use reflect to iterate over pointers
	if err := c.API.Verify(); err != nil {
		return err
	}
	if err := c.Web.Verify(); err != nil {
		return err
	}
	if err := c.Storage.Verify(); err != nil {
		return err
	}
	if err := c.Metrics.Verify(); err != nil {
		return err
	}
	if err := c.Logging.Verify(); err != nil {
		return err
	}
	return nil
}

// FromYAMLFile builds config structure from path
func FromYAMLFile(path string) (*Config, error) {
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
