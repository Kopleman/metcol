package metrics

import (
	"errors"
	"github.com/Kopleman/metcol/internal/store"
	"strconv"
	"strings"
)

type IMetrics interface {
	SetMetric(metricType MetricType, name string, value string) error
	GetValueAsString(metricType MetricType, name string) (string, error)
	GetAllValuesAsString() (map[string]string, error)
}

func (m *Metrics) buildStoreKey(name string, metricType MetricType) string {
	return name + "-" + string(metricType)
}

func (m *Metrics) parseStoreKey(key string) (string, MetricType, error) {
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return "", UnknownMetricType, ErrStoreKeyParse
	}

	typeAsString := parts[1]
	switch typeAsString {
	case string(CounterMetricType):
		return parts[0], CounterMetricType, nil
	case string(GougeMetricType):
		return parts[0], GougeMetricType, nil
	default:
		return parts[0], UnknownMetricType, ErrUnknownMetricType
	}
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

func (m *Metrics) GetValueAsString(metricType MetricType, name string) (string, error) {
	storeKey := m.buildStoreKey(name, metricType)
	value, err := m.store.Read(storeKey)
	if err != nil {
		return "", err
	}

	return m.convertMetricValueToString(metricType, value)
}

func (m *Metrics) GetAllValuesAsString() (map[string]string, error) {
	dataToReturn := make(map[string]string)
	allMetrics := m.store.GetAll()

	for metricKey, metricValue := range allMetrics {
		metricName, metricType, err := m.parseStoreKey(metricKey)
		if err != nil {
			return dataToReturn, err
		}
		valueAsString, err := m.convertMetricValueToString(metricType, metricValue)
		if err != nil {
			return dataToReturn, err
		}
		dataToReturn[metricName] = valueAsString
	}

	return dataToReturn, nil
}

func (m *Metrics) convertMetricValueToString(metricType MetricType, value any) (string, error) {
	switch metricType {
	case CounterMetricType:
		typedValue, ok := value.(int64)
		if !ok {
			return "", ErrCounterValueParse
		}
		return strconv.FormatInt(typedValue, 10), nil
	case GougeMetricType:
		typedValue, ok := value.(float64)
		if !ok {
			return "", ErrGougeValueParse
		}
		return strconv.FormatFloat(typedValue, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
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
