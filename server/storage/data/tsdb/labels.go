package tsdb

import (
	"github.com/prometheus/tsdb/labels"
)

const (
	sessionLabelName    = "session"
	metaLabelName       = "meta"
	metricTypeLabelName = "metric_type"
)

func allocBytesLabel() labels.Label {
	return labels.Label{Name: metricTypeLabelName, Value: "AllocBytes"}
}
func allocObjectsLabel() labels.Label {
	return labels.Label{Name: metricTypeLabelName, Value: "AllocObjects"}
}
func freeBytesLabel() labels.Label {
	return labels.Label{Name: metricTypeLabelName, Value: "FreeBytes"}
}
func freeObjectsLabel() labels.Label {
	return labels.Label{Name: metricTypeLabelName, Value: "FreeObjects"}
}

type measurementInfo struct {
	Labels labels.Labels
	Value  float64
}

type measurementsInfo []measurementInfo
