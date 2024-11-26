package metricscollector

import "github.com/Kopleman/metcol/internal/metrics"

type MetricItem struct {
	value      string
	metricType metrics.MetricType
}
