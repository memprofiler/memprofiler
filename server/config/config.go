package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config is a top-level structure with all server settings
type Config struct {
	Backend         *BackendConfig         `yaml:"backend"`
	Frontend        *FrontendConfig        `yaml:"frontend"`
	Metrics         *MetricsConfig         `yaml:"metrics"`
	Logging         *LoggingConfig         `yaml:"logging"`
	DataStorage     *DataStorageConfig     `yaml:"data_storage"`
	MetadataStorage *MetadataStorageConfig `yaml:"metadata_storage"`
}

// Verify checks config
func (c *Config) Verify() error {
	// TODO: use reflect to iterate over pointers
	if err := c.Backend.Verify(); err != nil {
		return errors.Wrap(err, "backend")
	}
	if err := c.Frontend.Verify(); err != nil {
		return errors.Wrap(err, "frontend")
	}
	if err := c.Metrics.Verify(); err != nil {
		return errors.Wrap(err, "metrics")
	}
	if err := c.Logging.Verify(); err != nil {
		return errors.Wrap(err, "logging")
	}
	if err := c.DataStorage.Verify(); err != nil {
		return errors.Wrap(err, "data_storage")
	}
	if err := c.MetadataStorage.Verify(); err != nil {
		return errors.Wrap(err, "metadata_storage")
	}

	return nil
}

// FromYAMLFile builds config structure from path
func FromYAMLFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}
	var c Config
	if err = yaml.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrap(err, "parse config file")
	}

	if err = c.Verify(); err != nil {
		return nil, errors.Wrap(err, "verify config file")
	}

	return &c, nil
}
