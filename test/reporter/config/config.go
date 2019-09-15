package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/utils"
)

// Config is a configuration for utility application that provides
// some memory consumption reports to the application
type Config struct {
	// Memory consumption scenario
	Scenario *Scenario `yaml:"scenario"`
	// Memprofiler client config
	Client *client.Config `yaml:"client"`
}

// Scenario configures application behavior in terms of memory consumption
type Scenario struct {
	// Sequence of elementary steps to emulate some memory strategy.
	// Will be played in a loop (like a circle)
	Steps []*Step `yaml:"steps"`
}

// Step is an minimal element describing memory consumption strategy behavior
type Step struct {
	// How much memory should be added / dropped at once
	MemoryDelta int `yaml:"memory_delta"`
	// Pause duration after memory allocation / freeing
	Wait utils.Duration `yaml:"wait"`
}

// FromYAMLFile builds config structure from YAML formatted file
func FromYAMLFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	var c Config
	if err = yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
