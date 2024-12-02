package metrics

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/server/store"
)

type IMetrics interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetAllValuesAsString() (map[string]string, error)
}

func (m *Metrics) buildStoreKey(name string, metricType common.MetricType) string {
	return name + "-" + string(metricType)
}

func (m *Metrics) parseStoreKey(key string) (string, common.MetricType, error) {
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return "", common.UnknownMetricType, ErrStoreKeyParse
	}

	typeAsString := parts[1]
	switch typeAsString {
	case string(common.CounterMetricType):
		return parts[0], common.CounterMetricType, nil
	case string(common.GougeMetricType):
		return parts[0], common.GougeMetricType, nil
	default:
		return parts[0], common.UnknownMetricType, ErrUnknownMetricType
	}
}

func (m *Metrics) SetGauge(name string, value float64) error {
	storeKey := m.buildStoreKey(name, common.GougeMetricType)

	_, err := m.store.Read(storeKey)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return m.store.Create(storeKey, value)
		}

		return err
	}

	return m.store.Update(storeKey, value)
}

func (m *Metrics) SetCounter(name string, value int64) error {
	storeKey := m.buildStoreKey(name, common.CounterMetricType)

	counterValue, err := m.store.Read(storeKey)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return m.store.Create(storeKey, value)
		}

		return err
	}

	parsedValue, ok := counterValue.(int64)

	if !ok {
		return ErrCounterValueParse
	}

	return m.store.Update(storeKey, parsedValue+value)
}

func (m *Metrics) SetMetric(metricType common.MetricType, name string, value string) error {
	switch metricType {
	case common.CounterMetricType:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrValueParse
		}
		return m.SetCounter(name, parsedValue)
	case common.GougeMetricType:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrValueParse
		}
		return m.SetGauge(name, parsedValue)
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) GetValueAsString(metricType common.MetricType, name string) (string, error) {
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

func (m *Metrics) convertMetricValueToString(metricType common.MetricType, value any) (string, error) {
	switch metricType {
	case common.CounterMetricType:
		typedValue, ok := value.(int64)
		if !ok {
			return "", ErrCounterValueParse
		}
		return strconv.FormatInt(typedValue, 10), nil
	case common.GougeMetricType:
		typedValue, ok := value.(float64)
		if !ok {
			return "", ErrGougeValueParse
		}
		return strconv.FormatFloat(typedValue, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
	}
}

func ParseMetricType(typeAsString string) (common.MetricType, error) {
	switch typeAsString {
	case string(common.CounterMetricType):
		return common.CounterMetricType, nil
	case string(common.GougeMetricType):
		return common.GougeMetricType, nil
	default:
		return common.UnknownMetricType, ErrUnknownMetricType
	}

}

type Store interface {
	Create(key string, value any) error
	Read(key string) (any, error)
	Update(key string, value any) error
	GetAll() map[string]any
}

type Metrics struct {
	store Store
}

func NewMetrics(s Store) *Metrics {
	return &Metrics{store: s}
}
