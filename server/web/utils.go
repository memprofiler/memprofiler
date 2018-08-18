package web

import (
	"encoding/json"
	"io"

	"math"
	"reflect"

	"github.com/vitalyisaev2/memprofiler/schema"
)

const (
	prefix = ""
	indent = "    "
)

// newJSONEncoder makes new encoder that creates pretty-printed JSON files
func newJSONEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	encoder.SetIndent(prefix, indent)
	return encoder
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
