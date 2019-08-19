package config

import (
	"github.com/rs/zerolog"
)

// LoggingConfig contains logging settings
type LoggingConfig struct {
	LevelString string        `yaml:"level"`
	Level       zerolog.Level `yaml:"-"`
}

// Verify checks config
func (c *LoggingConfig) Verify() error {
	var err error
	c.Level, err = zerolog.ParseLevel(c.LevelString)
	return err
}
