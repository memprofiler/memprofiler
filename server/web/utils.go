package web

import (
	"math"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/vitalyisaev2/memprofiler/schema"
)

const (
	indent = "    "
)

// newJSONMarshaler makes new marshaler that creates pretty-printed JSON files
func newJSONMarshaler() *jsonpb.Marshaler {
	marshaler := &jsonpb.Marshaler{
		Indent:       indent,
		EmitDefaults: true,
	}
	return marshaler
}

// sanitizeLocationMetricsForJSON drops values that are not available in JSON format
func sanitizeLocationMetricsForJSON(lms []*schema.LocationMetrics) {
	for _, lm := range lms {
		sanitizeHeapConsumptionRatesForJSON(lm.Average)
		sanitizeHeapConsumptionRatesForJSON(lm.Recent)
	}
}

// sanitizeHeapConsumptionRatesForJSON drops NaN values
func sanitizeHeapConsumptionRatesForJSON(hcr *schema.HeapConsumptionRates) {
	if hcr == nil {
		return
	}
	v := reflect.ValueOf(hcr).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Float64 {
			if math.IsNaN(f.Float()) {
				f.SetFloat(0)
			}
		}
	}
}
