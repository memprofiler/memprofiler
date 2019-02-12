package config

import "fmt"

// StorageConfig contains service storage settings
type StorageConfig struct {
	Filesystem *FilesystemStorageConfig `yaml:"filesystem"`
}

// Verify verifies config
func (c *StorageConfig) Verify() error {
	if c.Filesystem == nil {
		return fmt.Errorf("empty filesystem section")
	}

	return c.Filesystem.Verify()
}

// FilesystemStorageConfig contains settings of Filesystem-based storage
type FilesystemStorageConfig struct {
	// DataDir containes path to root directory to keep measurement data
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
