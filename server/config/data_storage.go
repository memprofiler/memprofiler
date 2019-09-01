package config

import (
	"fmt"

	"github.com/pkg/errors"
)

// DataStorageConfig contains various options for data storage
type DataStorageConfig struct {
	Filesystem      *FilesystemStorageConfig `yaml:"filesystem"`
	TSDB            *TSDBStorageConfig       `yaml:"tsdb"`
	dataStorageType DataStorageType
}

type DataStorageType int

const (
	FilesystemDataStorage DataStorageType = iota + 1
	TSDBDataStorage
)

func (c *DataStorageConfig) Verify() error {
	if c == nil {
		return fmt.Errorf("empty data storage config")
	}
	if (c.Filesystem == nil && c.TSDB == nil) || (c.Filesystem != nil && c.TSDB != nil) {
		return fmt.Errorf("you should provide either 'filesystem' or 'tsdb' configs")
	}
	if c.Filesystem != nil {
		c.dataStorageType = FilesystemDataStorage
		if err := c.Filesystem.Verify(); err != nil {
			return errors.Wrap(err, "filesystem storage")
		}
	}
	if c.TSDB != nil {
		c.dataStorageType = TSDBDataStorage
		if err := c.TSDB.Verify(); err != nil {
			return errors.Wrap(err, "tsdb storage")
		}
	}
	return nil
}

func (c DataStorageConfig) Type() DataStorageType { return c.dataStorageType }

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
		return fmt.Errorf("empty data_dir")
	}
	return nil
}

// TSDBStorageConfig contains settings of Prometheus TSDB-based storage
type TSDBStorageConfig struct {
	// DataDir contains path to root directory to keep measurement data
	DataDir string `yaml:"data_dir"`
}

// Verify verifies config
func (c *TSDBStorageConfig) Verify() error {
	if c.DataDir == "" {
		return fmt.Errorf("empty data_dir")
	}
	return nil
}
