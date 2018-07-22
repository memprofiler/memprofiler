package config

import "github.com/sirupsen/logrus"

// LoggingConfig contains logging settings
type LoggingConfig struct {
	LevelString string       `yaml:"level"`
	Level       logrus.Level `yaml:"-"`
}

func (c *LoggingConfig) Verify() error {
	var err error
	c.Level, err = logrus.ParseLevel(c.LevelString)
	return err
}
