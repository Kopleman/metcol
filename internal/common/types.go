package common

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GougeMetricType   MetricType = "gauge"
	UnknownMetricType MetricType = "unknown"
)

var (
	metricTypeName = map[MetricType]string{
		CounterMetricType: "counter",
		GougeMetricType:   "gauge",
		UnknownMetricType: "unknown",
	}
	metricTypeValue = map[string]MetricType{
		"counter": CounterMetricType,
		"gauge":   GougeMetricType,
		"unknown": UnknownMetricType,
	}
)

func (mt MetricType) String() string {
	return metricTypeName[mt]
}

func StringToMetricType(s string) (MetricType, bool) {
	res, ok := metricTypeValue[s]
	return res, ok
}
