package filesystem

import (
	"github.com/prometheus/tsdb/labels"
)

const (
	SessionLabelName    = "session"
	MetaLabelName       = "meta"
	MetricTypeLabelName = "metric_type"
)

var (
	AllocBytesLabel   = labels.Label{Name: MetricTypeLabelName, Value: "AllocBytes"}
	AllocObjectsLabel = labels.Label{Name: MetricTypeLabelName, Value: "AllocObjects"}
	FreeBytesLabel    = labels.Label{Name: MetricTypeLabelName, Value: "FreeBytes"}
	FreeObjectsLabel  = labels.Label{Name: MetricTypeLabelName, Value: "FreeObjects"}
)
