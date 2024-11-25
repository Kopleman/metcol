package metrics

import (
	"errors"
	"github.com/Kopleman/metcol/internal/store"
	"strconv"
)

type IMetrics interface {
	SetMetric(metricType MetricType, name string, value string) error
	Get(metricType MetricType, name string) (any, error)
}

func (m *Metrics) buildStoreKey(name string, metricType MetricType) string {
	return name + "-" + string(metricType)
}

func (m *Metrics) SetGauge(name string, value float64) error {
	storeKey := m.buildStoreKey(name, GougeMetricType)

	_, err := m.store.Read(storeKey)

	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	if err != nil && errors.Is(err, store.ErrNotFound) {
		return m.store.Create(storeKey, value)
	}

	return m.store.Update(storeKey, value)
}

func (m *Metrics) SetCounter(name string, value int64) error {
	storeKey := m.buildStoreKey(name, CounterMetricType)

	counterValue, err := m.store.Read(storeKey)

	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	if err != nil && errors.Is(err, store.ErrNotFound) {
		return m.store.Create(storeKey, value)
	}

	parsedValue, ok := counterValue.(int64)

	if !ok {
		return ErrCounterValueParse
	}

	return m.store.Update(storeKey, parsedValue+value)
}

func (m *Metrics) SetMetric(metricType MetricType, name string, value string) error {
	switch metricType {
	case CounterMetricType:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrValueParse
		}
		return m.SetCounter(name, parsedValue)
	case GougeMetricType:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrValueParse
		}
		return m.SetGauge(name, parsedValue)
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) Get(metricType MetricType, name string) (any, error) {
	storeKey := m.buildStoreKey(name, metricType)
	return m.store.Read(storeKey)
}

func ValidateMetricsValue(metricType MetricType, value any) bool {
	switch metricType {
	case CounterMetricType:
		_, ok := value.(int64)
		return ok
	case GougeMetricType:
		_, ok := value.(float64)
		return ok
	default:
		return false
	}
}

func ParseMetricType(typeAsString string) (MetricType, error) {
	switch typeAsString {
	case string(CounterMetricType):
		return CounterMetricType, nil
	case string(GougeMetricType):
		return GougeMetricType, nil
	default:
		return UnknownMetricType, ErrUnknownMetricType
	}

}

type Metrics struct {
	store store.IStore
}

func NewMetrics(s store.IStore) IMetrics {
	return &Metrics{store: s}
}
