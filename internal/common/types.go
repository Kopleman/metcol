package common

// MetricType type for metrics.
type MetricType string

const (
	CounterMetricType MetricType = "counter" // counter type.
	GaugeMetricType   MetricType = "gauge"   // gauge type.
	UnknownMetricType MetricType = "unknown" // unknown type
)

var (
	metricTypeName = map[MetricType]string{
		CounterMetricType: "counter",
		GaugeMetricType:   "gauge",
		UnknownMetricType: "unknown",
	}
	metricTypeValue = map[string]MetricType{
		"counter": CounterMetricType,
		"gauge":   GaugeMetricType,
		"unknown": UnknownMetricType,
	}
)

// String converts type to plain string.
func (mt MetricType) String() string {
	return metricTypeName[mt]
}

// StringToMetricType convert string to metric type.
func StringToMetricType(s string) (MetricType, bool) {
	res, ok := metricTypeValue[s]
	return res, ok
}
