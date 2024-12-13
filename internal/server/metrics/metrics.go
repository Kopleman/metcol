package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/server/store"
)

func (m *Metrics) buildStoreKey(name string, metricType common.MetricType) string {
	return name + "-" + string(metricType)
}

const metricsStoreKeyPartsNum = 2

func (m *Metrics) parseStoreKey(key string) (string, common.MetricType, error) {
	parts := strings.Split(key, "-")
	if len(parts) != metricsStoreKeyPartsNum {
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

func (m *Metrics) SetGauge(name string, value float64) (*float64, error) {
	storeKey := m.buildStoreKey(name, common.GougeMetricType)

	_, err := m.store.Read(storeKey)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			storeErr := m.store.Create(storeKey, value)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create gauge metric '%s': %w", storeKey, err)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read gauge metric '%s': %w", storeKey, err)
	}

	updateErr := m.store.Update(storeKey, value)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update gauge metric '%s': %w", storeKey, err)
	}

	return &value, nil
}

func (m *Metrics) SetCounter(name string, value int64) (*int64, error) {
	storeKey := m.buildStoreKey(name, common.CounterMetricType)

	counterValue, err := m.store.Read(storeKey)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			storeErr := m.store.Create(storeKey, value)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create counter metric '%s': %w", storeKey, err)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read counter metric '%s': %w", storeKey, err)
	}

	parsedValue, ok := counterValue.(int64)

	if !ok {
		return nil, ErrCounterValueParse
	}

	newValue := parsedValue + value
	updateErr := m.store.Update(storeKey, newValue)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update counter metric '%s': %w", storeKey, err)
	}

	return &newValue, nil
}

func (m *Metrics) SetMetric(metricType common.MetricType, name string, value string) error {
	switch metricType {
	case common.CounterMetricType:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrValueParse
		}
		_, err = m.SetCounter(name, parsedValue)
		return err
	case common.GougeMetricType:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrValueParse
		}
		_, err = m.SetGauge(name, parsedValue)
		return err
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) SetMetricByDto(d *dto.MetricDto) error {
	switch d.MType {
	case common.CounterMetricType:
		if d.Delta == nil {
			return ErrValueParse
		}
		newDelta, err := m.SetCounter(d.ID, *d.Delta)
		if err != nil {
			return err
		}
		d.Delta = newDelta
		return nil
	case common.GougeMetricType:
		if d.Value == nil {
			return ErrValueParse
		}
		newValue, err := m.SetGauge(d.ID, *d.Value)
		if err != nil {
			return err
		}
		d.Value = newValue
		return nil
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) GetValueAsString(metricType common.MetricType, name string) (string, error) {
	storeKey := m.buildStoreKey(name, metricType)
	value, err := m.store.Read(storeKey)
	if err != nil {
		return "", fmt.Errorf("failed to read metric '%s': %w", storeKey, err)
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
