package metrics

import (
	"reflect"

	"gonum.org/v1/gonum/stat"

	"github.com/vitalyisaev2/memprofiler/schema"
)

// locationData contains collection of time series with different heap metrics;
// some of fields are public to make the work with reflection little bit easy
type locationData struct {
	AllocBytes   []float64
	AllocObjects []float64
	FreeBytes    []float64
	FreeObjects  []float64
	InUseBytes   []float64
	InUseObjects []float64
	Timestamps   []float64
	window       int
	callStack    *schema.Callstack
}

// registerMeasurement appends new measurement to the process
func (ld *locationData) registerMeasurement(timestamp float64, mu *schema.MemoryUsage) {

	// shift series if the maximum capacity is achieved
	if len(ld.AllocBytes) == ld.window {
		ld.AllocObjects = ld.AllocObjects[:ld.window-1]
		ld.AllocBytes = ld.AllocBytes[:ld.window-1]
		ld.FreeObjects = ld.FreeObjects[:ld.window-1]
		ld.FreeBytes = ld.FreeBytes[:ld.window-1]
		ld.InUseObjects = ld.InUseObjects[:ld.window-1]
		ld.InUseBytes = ld.InUseBytes[:ld.window-1]
		ld.Timestamps = ld.Timestamps[:ld.window-1]
	}

	// add required data
	ld.AllocObjects = append(ld.AllocObjects, float64(mu.AllocObjects))
	ld.AllocBytes = append(ld.AllocBytes, float64(mu.AllocBytes))
	ld.FreeObjects = append(ld.FreeObjects, float64(mu.FreeObjects))
	ld.FreeBytes = append(ld.FreeBytes, float64(mu.FreeBytes))
	ld.InUseObjects = append(ld.InUseObjects, float64(mu.AllocObjects)-float64(mu.FreeObjects))
	ld.InUseBytes = append(ld.InUseBytes, float64(mu.AllocBytes)-float64(mu.FreeBytes))
	ld.Timestamps = append(ld.Timestamps, timestamp)
}

// computeMetrics performs stats computations for every stored time series;
// the timestamps are shared between all session members and stored out there
func (ld *locationData) computeMetrics() *schema.LocationMetrics {

	rates := &schema.HeapConsumptionRates{}
	ldValue := reflect.Indirect(reflect.ValueOf(ld))
	ratesValue := reflect.Indirect(reflect.ValueOf(rates))

	for i := 0; i < ldValue.NumField(); i++ {
		ldField := ldValue.Field(i)

		fieldName := ldValue.Type().Field(i).Name

		// estimate regression parameters for every time series
		if fieldName != "Timestamps" &&
			ldField.Kind() == reflect.Slice &&
			ldField.Type().Elem().Kind() == reflect.Float64 {
			slope := computeSlope(ld.Timestamps, ldField.Interface().([]float64))
			ratesField := ratesValue.FieldByName(fieldName)
			ratesField.SetFloat(slope)
		}
	}

	return &schema.LocationMetrics{
		Rates:     rates,
		Callstack: ld.callStack,
	}
}

// computeSlope computes the slope of linear regression equation,
// which is equal to rate [units per second], or the first time derivative
func computeSlope(tstamps, values []float64) float64 {
	x := tstamps
	if len(tstamps) != len(values) {
		x = tstamps[len(values):]
	}
	_, slope := stat.LinearRegression(x, values, nil, false)
	return slope
}

func newLocationData(callStack *schema.Callstack, window int) *locationData {
	return &locationData{
		callStack:    callStack,
		window:       window,
		AllocBytes:   make([]float64, 0, window),
		AllocObjects: make([]float64, 0, window),
		FreeBytes:    make([]float64, 0, window),
		FreeObjects:  make([]float64, 0, window),
		InUseBytes:   make([]float64, 0, window),
		InUseObjects: make([]float64, 0, window),
	}
}
