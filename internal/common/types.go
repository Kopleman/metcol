package common

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GougeMetricType   MetricType = "gauge"
	UnknownMetricType MetricType = "unknown"
)
