package web

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/memprofiler/schema"
)

func TestSanitize(t *testing.T) {
	lms := []*schema.LocationMetrics{
		{
			Average: &schema.HeapConsumptionRates{
				AllocObjectsRate: math.NaN(),
				AllocBytesRate:   1.0,
			},
		},
	}

	sanitizeLocationMetricsForJSON(lms)

	assert.Equal(t, float64(0), lms[0].Average.AllocObjectsRate)
	assert.Equal(t, float64(1.0), lms[0].Average.AllocBytesRate)
}
