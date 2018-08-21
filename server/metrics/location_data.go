package metrics

import (
	"reflect"

	"github.com/vitalyisaev2/memprofiler/schema"
	"gonum.org/v1/gonum/stat"
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
	window       int
	callStack    *schema.CallStack
}

// registerMeasurement appends new measurement to the process
func (ld *locationData) registerMeasurement(mu *schema.MemoryUsage) {

	// shift series if the maximum capacity is achieved
	if len(ld.AllocBytes) == ld.window {
		ld.AllocObjects = ld.AllocObjects[:ld.window-1]
		ld.AllocBytes = ld.AllocBytes[:ld.window-1]
		ld.FreeObjects = ld.FreeBytes[:ld.window-1]
		ld.FreeBytes = ld.FreeBytes[:ld.window-1]
		ld.InUseObjects = ld.InUseBytes[:ld.window-1]
		ld.InUseBytes = ld.InUseBytes[:ld.window-1]
	}

	// add required data
	ld.AllocObjects = append(ld.AllocObjects, float64(mu.AllocObjects))
	ld.AllocBytes = append(ld.AllocBytes, float64(mu.AllocBytes))
	ld.FreeObjects = append(ld.FreeObjects, float64(mu.FreeObjects))
	ld.FreeBytes = append(ld.FreeBytes, float64(mu.FreeBytes))
	ld.InUseObjects = append(ld.InUseObjects, float64(mu.AllocObjects)-float64(mu.FreeObjects))
	ld.InUseBytes = append(ld.InUseObjects, float64(mu.AllocBytes)-float64(mu.FreeBytes))
}

// computeMetrics performs stats computations for every stored time series;
// the timestamps are shared between all session members and stored out there
func (ld *locationData) computeMetrics(tstamps []float64) *schema.LocationMetrics {

	var rates schema.HeapConsumptionRates

	ldValue := reflect.Indirect(reflect.ValueOf(ld))
	ratesValue := reflect.ValueOf(rates)

	for i := 0; i < ldValue.NumField(); i++ {
		ldField := ldValue.Field(i)

		// if field type is []float64, compute the metrics
		if ldField.Kind() == reflect.Slice && ldField.Type().Elem().Kind() == reflect.Float64 {
			slope := computeSlope(tstamps, ldField.Interface().([]float64))
			ratesField := ratesValue.FieldByName(ldValue.Type().Field(i).Name)
			ratesField.SetFloat(slope)
		}
	}

	return &schema.LocationMetrics{
		Rates:     &rates,
		CallStack: ld.callStack,
	}
}

// computeSlope computes the slope of linear regression equation,
// which is equal to rate [units per second] or the first time derivative
func computeSlope(tstamps, values []float64) float64 {
	_, slope := stat.LinearRegression(tstamps, values, nil, false)
	return slope
}

func newLocationData(callStack *schema.CallStack, window int) *locationData {
	return &locationData{
		window:    window,
		callStack: callStack,
	}
}
