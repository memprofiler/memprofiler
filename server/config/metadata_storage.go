package config

import "fmt"

// MetadataStorageConfig contains config for session metadata storage
type MetadataStorageConfig struct {
	// DataDir - directory with data
	DataDir string `yaml:"data_dir"`
}

// Verify checks config
func (c *MetadataStorageConfig) Verify() error {
	if c == nil {
		return fmt.Errorf("empty metadata storage config")
	}
	if c.DataDir == "" {
		return fmt.Errorf("empty data_dir")
	}
	return nil
}
