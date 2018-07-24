package collector

import (
	"sync"

	"github.com/vitalyisaev2/memprofiler/schema"
)

type defaultCollector struct {
	mutex *sync.RWMutex
	stats *overallStats
}

func (c *defaultCollector) RegisterMeasurement(
	desc *schema.ServiceDescription,
	mm *schema.Measurement,
) error {

	// obtain or initialize top level data structures
	c.mutex.Lock()

	ss, exists := c.stats.services[desc.GetType()]
	if !exists {
		ss = newServiceStats()
		c.stats.services[desc.GetType()] = ss
	}

	is, exists := ss.instances[desc.GetInstance()]
	if !exists {
		is = newInstanceStats()
		ss.instances[desc.GetInstance()] = is
	}

	c.mutex.Unlock()

	return nil
}

func New() Service {
	return &defaultCollector{
		stats: newOverallStats(),
		mutex: &sync.RWMutex{},
	}
}
