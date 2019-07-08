package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is a top-level structure with all server settings
type Config struct {
	Backend     *BackendConfig           `yaml:"backend"`
	Frontend    *FrontendConfig          `yaml:"frontend"`
	Metrics     *MetricsConfig           `yaml:"metrics"`
	Logging     *LoggingConfig           `yaml:"logging"`
	StorageType string                   `yaml:"storage_type"`
	Filesystem  *FilesystemStorageConfig `yaml:"filesystem"`
	TSDB        *TSDBStorageConfig       `yaml:"tsdb"`
}

// Verify checks config
func (c *Config) Verify() error {
	// TODO: use reflect to iterate over pointers
	if err := c.Backend.Verify(); err != nil {
		return err
	}
	if err := c.Frontend.Verify(); err != nil {
		return err
	}
	if err := c.Metrics.Verify(); err != nil {
		return err
	}
	if err := c.Logging.Verify(); err != nil {
		return err
	}

	switch c.StorageType {
	case StorageTypeTSDB:
		if err := c.TSDB.Verify(); err != nil {
			return err
		}
	case StorageTypeFilesystem:
		if err := c.Filesystem.Verify(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected storage type")
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
