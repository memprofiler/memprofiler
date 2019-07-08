package config

import "fmt"

// FilesystemStorageConfig contains settings of Filesystem-based storage
type FilesystemStorageConfig struct {
	// DataDir contains path to root directory to keep measurement data
	DataDir string `yaml:"data_dir"`
	// SyncWrite enables fsync after every write
	SyncWrite bool `yaml:"sync_write"`
}

// Verify verifies config
func (c *FilesystemStorageConfig) Verify() error {
	if c.DataDir == "" {
		return fmt.Errorf("invalid FilesystemStorageConfig.DataDir")
	}
	return nil
}

// TSDBStorageConfig contains settings of TSDB-based storage
type TSDBStorageConfig struct {
	// DataDir contains path to root directory to keep measurement data
	DataDir string `yaml:"data_dir"`
}

// Verify verifies config
func (c *TSDBStorageConfig) Verify() error {
	if c.DataDir == "" {
		return fmt.Errorf("invalid FilesystemStorageConfig.DataDir")
	}
	return nil
}
