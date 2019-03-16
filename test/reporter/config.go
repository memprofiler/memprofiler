package main

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/memprofiler/memprofiler/client"
)

// Config is a configuration for utility application that provides
// some memory consumption reports to the application
type Config struct {
	// Memory consumption scenario
	Scenario *Scenario `json:"scenario"`
	// Memprofiler client config
	Client *client.Config `json:"client"`
}

// Scenario configures application behavior in terms of memory consumption
type Scenario struct {
	// Sequence of elementary steps to emulate some memory strategy.
	// Will be played in a loop (like a circle)
	Steps []*Step `json:"step"`
}

// Step is an minimal element describing memory consumption strategy behaviour
type Step struct {
	// How much memory should be added / dropped at once
	MemoryDelta int `json:"memory_delta"`
	// Pause duration after memory allocation / freeing
	Wait time.Duration `json:"wait"`
}

// FromYAMLFile builds config structure from YAML formatted file
func FromYAMLFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err = yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
