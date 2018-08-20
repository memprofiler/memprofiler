package metrics

import "github.com/vitalyisaev2/memprofiler/schema"

type locationData struct {
	allocBytes   []float64
	allocObjects []float64
	freeBytes    []float64
	freeObjects  []float64
	inUseBytes   []float64
	inUseObjects []float64
	window       int
}

func (ld *locationData) registerMeasurement(mu *schema.MemoryUsage) {
	if len(ld.allocBytes) == ld.window {
		ld.allocObjects = ld.allocObjects[:ld.window-1]
		ld.allocBytes = ld.allocBytes[:ld.window-1]
		ld.freeObjects = ld.freeBytes[:ld.window-1]
		ld.freeBytes = ld.freeBytes[:ld.window-1]
		ld.inUseObjects = ld.inUseBytes[:ld.window-1]
		ld.inUseBytes = ld.inUseBytes[:ld.window-1]
	}

	ld.allocObjects = append(ld.allocObjects, float64(mu.AllocObjects))
	ld.allocBytes = append(ld.allocBytes, float64(mu.AllocBytes))
	ld.freeObjects = append(ld.freeObjects, float64(mu.FreeObjects))
	ld.freeBytes = append(ld.freeBytes, float64(mu.FreeBytes))
	ld.inUseObjects = append(ld.inUseObjects, float64(mu.AllocObjects)-float64(mu.FreeObjects))
	ld.inUseBytes = append(ld.inUseObjects, float64(mu.AllocBytes)-float64(mu.FreeBytes))
}
