package metrics

import (
	"reflect"
	"sort"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/vitalyisaev2/memprofiler/utils"

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
	Timestamps   []time.Time
	lifetime     time.Duration // equals to the longest averaging window available
	callStack    *schema.Callstack
}

// registerMeasurement appends new measurement to the process
func (ld *locationData) registerMeasurement(timestamp time.Time, mu *schema.MemoryUsage) {

	// check if there are outdated records
	threshold := time.Now().Add(-1 * ld.lifetime)
	edge := 0
	for i, timestamp := range ld.Timestamps {
		if timestamp.Before(threshold) {
			edge = i
		} else {
			break
		}
	}

	// shift series if data TTL is reached
	if edge != 0 {
		ld.AllocObjects = ld.AllocObjects[edge:]
		ld.AllocBytes = ld.AllocBytes[edge:]
		ld.FreeObjects = ld.FreeObjects[edge:]
		ld.FreeBytes = ld.FreeBytes[edge:]
		ld.InUseObjects = ld.InUseObjects[edge:]
		ld.InUseBytes = ld.InUseBytes[edge:]
		ld.Timestamps = ld.Timestamps[edge:]
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
func (ld *locationData) computeMetrics(spans []time.Duration) *schema.LocationMetrics {

	// x axis values
	timestampFloats := timestampsToFloats(ld.Timestamps)

	rates := make([]*schema.MemoryUtilizationRate, len(spans))
	result := &schema.LocationMetrics{Callstack: ld.callStack, Rates: rates}

	for i, span := range spans {
		rate := ld.computeMetricsForSpan(span, timestampFloats)
		rates[i] = rate
	}
	return result
}

// computeMetricsForSpan performs stats computation for a particular time span
func (ld *locationData) computeMetricsForSpan(
	span time.Duration,
	timestampFloats []float64,
) *schema.MemoryUtilizationRate {

	threshold := utils.TimeToFloat64(time.Now().Add(-1 * span))
	ix := sort.SearchFloat64s(timestampFloats, threshold)

	result := &schema.MemoryUtilizationRate{
		Values: &schema.MemoryUtilizationRate_Values{},
		Span:   ptypes.DurationProto(span),
	}

	// walk through fields and cast
	src := reflect.Indirect(reflect.ValueOf(ld))
	dst := reflect.Indirect(reflect.ValueOf(result.Values))
	for i := 0; i < src.NumField(); i++ {

		dataField := src.Field(i)
		fieldName := src.Type().Field(i).Name

		// estimate regression parameters for every time series
		if fieldName != "Timestamps" &&
			dataField.Kind() == reflect.Slice &&
			dataField.Type().Elem().Kind() == reflect.Float64 {
			slope := computeSlope(timestampFloats[ix:], dataField.Interface().([]float64)[ix:])
			ratesField := dst.FieldByName(fieldName)
			ratesField.SetFloat(slope)
		}
	}

	return result
}

// timestampsToFloats converts array of timestamps to array of floats
func timestampsToFloats(timestamps []time.Time) []float64 {
	result := make([]float64, len(timestamps))
	for i, timestamp := range timestamps {
		result[i] = utils.TimeToFloat64(timestamp)
	}
	return result
}

// computeSlope computes the slope of linear regression equation,
// which is equal to rate [units per second], or the first time derivative
// TODO: is it a better way to estimate time series trend?
//  https://en.wikipedia.org/wiki/Spearman%27s_rank_correlation_coefficient
//  https://www.r-bloggers.com/trend-analysis-with-the-cox-stuart-test-in-r/
func computeSlope(timestamps, values []float64) float64 {
	x := timestamps
	_, slope := stat.LinearRegression(x, values, nil, false)
	return slope
}

func newLocationData(callStack *schema.Callstack, lifetime time.Duration) *locationData {
	return &locationData{
		callStack: callStack,
		lifetime:  lifetime,
	}
}
