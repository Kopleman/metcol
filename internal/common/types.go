package common

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GougeMetricType   MetricType = "gauge"
	UnknownMetricType MetricType = "unknown"
)

var (
	metricType_name = map[MetricType]string{
		CounterMetricType: "counter",
		GougeMetricType:   "gauge",
		UnknownMetricType: "unknown",
	}
	metricType_value = map[string]MetricType{
		"counter": CounterMetricType,
		"gauge":   GougeMetricType,
		"unknown": UnknownMetricType,
	}
)

func (mt MetricType) String() string {
	return metricType_name[mt]
}

func StringToMetricType(s string) (MetricType, bool) {
	res, ok := metricType_value[s]
	return res, ok
}
