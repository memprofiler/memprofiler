package metrics

import (
	"reflect"

	"gonum.org/v1/gonum/stat"

	"fmt"

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
	window       int
	callStack    *schema.CallStack
}

// registerMeasurement appends new measurement to the process
func (ld *locationData) registerMeasurement(mu *schema.MemoryUsage) {

	// shift series if the maximum capacity is achieved
	if len(ld.AllocBytes) == ld.window {
		ld.AllocObjects = ld.AllocObjects[:ld.window-1]
		ld.AllocBytes = ld.AllocBytes[:ld.window-1]
		ld.FreeObjects = ld.FreeObjects[:ld.window-1]
		ld.FreeBytes = ld.FreeBytes[:ld.window-1]
		ld.InUseObjects = ld.InUseObjects[:ld.window-1]
		ld.InUseBytes = ld.InUseBytes[:ld.window-1]
	}

	// add required data
	ld.AllocObjects = append(ld.AllocObjects, float64(mu.AllocObjects))
	ld.AllocBytes = append(ld.AllocBytes, float64(mu.AllocBytes))
	ld.FreeObjects = append(ld.FreeObjects, float64(mu.FreeObjects))
	ld.FreeBytes = append(ld.FreeBytes, float64(mu.FreeBytes))
	ld.InUseObjects = append(ld.InUseObjects, float64(mu.AllocObjects)-float64(mu.FreeObjects))
	ld.InUseBytes = append(ld.InUseBytes, float64(mu.AllocBytes)-float64(mu.FreeBytes))
}

// computeMetrics performs stats computations for every stored time series;
// the timestamps are shared between all session members and stored out there
func (ld *locationData) computeMetrics(tstamps []float64) *schema.LocationMetrics {

	rates := &schema.HeapConsumptionRates{}
	ldValue := reflect.Indirect(reflect.ValueOf(ld))
	ratesValue := reflect.Indirect(reflect.ValueOf(rates))

	for i := 0; i < ldValue.NumField(); i++ {
		ldField := ldValue.Field(i)

		// if field type is []float64, perform computation
		if ldField.Kind() == reflect.Slice && ldField.Type().Elem().Kind() == reflect.Float64 {
			slope := computeSlope(tstamps, ldField.Interface().([]float64))
			ratesField := ratesValue.FieldByName(ldValue.Type().Field(i).Name)
			ratesField.SetFloat(slope)
		}
	}

	return &schema.LocationMetrics{
		Rates:     rates,
		CallStack: ld.callStack,
	}
}

// computeSlope computes the slope of linear regression equation,
// which is equal to rate [units per second] or the first time derivative
func computeSlope(tstamps, values []float64) float64 {
	x := tstamps
	if len(tstamps) != len(values) {
		x = tstamps[len(values):]
	}
	fmt.Println(x, tstamps, values)
	_, slope := stat.LinearRegression(x, values, nil, false)
	return slope
}

func newLocationData(callStack *schema.CallStack, window int) *locationData {
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
