package common

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GaugeMetricType   MetricType = "gauge"
	UnknownMetricType MetricType = "unknown"
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

func (mt MetricType) String() string {
	return metricTypeName[mt]
}

func StringToMetricType(s string) (MetricType, bool) {
	res, ok := metricTypeValue[s]
	return res, ok
}
