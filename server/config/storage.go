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
	SyncWrite bool                          `yaml:"sync_write"`
	Cache     *FilesystemStorageCacheConfig `yaml:"cache"`
}

// Verify verifies config
func (c *FilesystemStorageConfig) Verify() error {
	if c.DataDir == "" {
		return fmt.Errorf("empty FilesystemStorageConfig.DataDir")
	}
	if c.Cache != nil {
		if err := c.Cache.Verify(); err != nil {
			return err
		}
	}
	return nil
}

// FilesystemStorageCacheConfig contains configuration for a caching layer within filesystem storage
type FilesystemStorageCacheConfig struct {
	// MaxTotalSize sets the upper limit for the total size (in-memory bytes) for values stored in cache
	MaxTotalSize int `yaml:"max_total_size"`
	// TTL sets lifetime for cached measurements (in seconds)
	TTL int `yaml:"ttl"`
	// GCFrequency sets the period for starting gc session (in seconds)
	GCFrequency int `yaml:"gc_frequency"`
	// GCMaxPause sets limitation for GC session timing in order
	// to prevent it from blocking cache for too long (in seconds)
	GCMaxPause int `yaml:"gc_max_pause"`
}

// Verify verifies config
func (c *FilesystemStorageCacheConfig) Verify() error {
	if c.MaxTotalSize == 0 {
		return fmt.Errorf("invalid FilesystemStorageCacheConfig.MaxTotalSize: %d", c.MaxTotalSize)
	}
	if c.TTL == 0 {
		return fmt.Errorf("invalid FilesystemStorageCacheConfig.TTL: %d", c.TTL)
	}
	if c.GCFrequency == 0 {
		return fmt.Errorf("invalid FilesystemStorageCacheConfig.GCFrequency: %d", c.GCFrequency)
	}
	if c.GCMaxPause == 0 {
		return fmt.Errorf("invalid FilesystemStorageCacheConfig.GCMaxPause: %d", c.GCMaxPause)
	}
	return nil
}
