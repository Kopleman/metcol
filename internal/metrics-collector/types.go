package metrics_collector

import "github.com/Kopleman/metcol/internal/metrics"

type MetricItem struct {
	value      string
	metricType metrics.MetricType
}
