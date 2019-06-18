package config

import "fmt"

// StorageConfig contains service storage settings
type StorageConfig struct {
	Filesystem *FilesystemStorageConfig `yaml:"filesystem"`
	TSDB       *TSDBStorageConfig       `yaml:"tsdb"`
}

// Verify verifies config
func (c *StorageConfig) Verify() error {
	if c.Filesystem == nil && c.TSDB == nil {
		return fmt.Errorf("no storage is set, use filesystem or tsdb")
	}

	if c.Filesystem != nil && c.TSDB != nil {
		return fmt.Errorf("please use only filesystem or tsdb")
	}

	if c.Filesystem != nil {
		err := c.Filesystem.Verify()
		if err != nil {
			return err
		}
	}

	if c.TSDB != nil {
		err := c.TSDB.Verify()
		if err != nil {
			return err
		}
	}

	return nil
}

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
