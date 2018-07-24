package collector

import "github.com/vitalyisaev2/memprofiler/schema"

type Service interface {
	RegisterMeasurement(desc *schema.ServiceDescription, mm *schema.Measurement) error
}
