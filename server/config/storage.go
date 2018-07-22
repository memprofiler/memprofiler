package config

import "fmt"

// StorageConfig contains service storage settings
type StorageConfig struct {
	Filesystem *FilesystemStorageConfig `yaml:"filesystem"`
}

func (c *StorageConfig) Verify() error {
	if c.Filesystem == nil {
		return fmt.Errorf("empty filesystem section")
	}

	return c.Filesystem.Verify()
}

// FilesystemStorageConfig contains settings of Filesystem-based storage
type FilesystemStorageConfig struct {
	DataDir   string `yaml:"data_dir"`
	SyncWrite bool   `yaml:"sync_write"`
}

func (c *FilesystemStorageConfig) Verify() error {
	if c.DataDir == "" {
		return fmt.Errorf("empty FilesystemStorageConfig.DataDir")
	}
	return nil
}
