package metrics

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/server/store_errors"
	"github.com/davecgh/go-spew/spew"
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

func (m *Metrics) SetGauge(ctx context.Context, name string, value float64) (*float64, error) {
	storeKey := m.buildStoreKey(name, common.GougeMetricType)

	_, err := m.store.Read(ctx, storeKey)

	metricDTO := &dto.MetricDTO{
		Delta: nil,
		Value: &value,
		ID:    name,
		MType: common.GougeMetricType,
	}
	if err != nil {
		if errors.Is(err, store_errors.ErrNotFound) {
			storeErr := m.store.Create(ctx, metricDTO)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create gauge metric '%s': %w", storeKey, err)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read gauge metric '%s': %w", storeKey, err)
	}

	updateErr := m.store.Update(ctx, metricDTO)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update gauge metric '%s': %w", storeKey, err)
	}

	return &value, nil
}

func (m *Metrics) SetCounter(ctx context.Context, name string, value int64) (*int64, error) {
	storeKey := m.buildStoreKey(name, common.CounterMetricType)

	existedCounter, err := m.store.Read(ctx, storeKey)

	if err != nil {
		if errors.Is(err, store_errors.ErrNotFound) {
			metricDTO := &dto.MetricDTO{
				Delta: &value,
				Value: nil,
				ID:    name,
				MType: common.CounterMetricType,
			}
			storeErr := m.store.Create(ctx, metricDTO)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create counter metric '%s': %w", storeKey, err)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read counter metric '%s': %w", storeKey, err)
	}

	if existedCounter.Delta == nil {
		return nil, ErrCounterValueParse
	}

	newValue := *existedCounter.Delta + value
	existedCounter.Delta = &newValue
	updateErr := m.store.Update(ctx, existedCounter)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update counter metric '%s': %w", storeKey, err)
	}

	return &newValue, nil
}

func (m *Metrics) SetMetric(ctx context.Context, metricType common.MetricType, name string, value string) error {
	switch metricType {
	case common.CounterMetricType:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrValueParse
		}
		_, err = m.SetCounter(ctx, name, parsedValue)
		return err
	case common.GougeMetricType:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrValueParse
		}
		_, err = m.SetGauge(ctx, name, parsedValue)
		return err
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) SetMetricByDto(ctx context.Context, d *dto.MetricDTO) error {
	switch d.MType {
	case common.CounterMetricType:
		if d.Delta == nil {
			return ErrValueParse
		}
		newDelta, err := m.SetCounter(ctx, d.ID, *d.Delta)
		if err != nil {
			return err
		}
		d.Delta = newDelta
		return nil
	case common.GougeMetricType:
		if d.Value == nil {
			return ErrValueParse
		}
		newValue, err := m.SetGauge(ctx, d.ID, *d.Value)
		if err != nil {
			return err
		}
		d.Value = newValue
		return nil
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error) {
	storeKey := m.buildStoreKey(name, metricType)
	value, err := m.store.Read(ctx, storeKey)
	if err != nil {
		return "", fmt.Errorf("failed to read metric '%s': %w", storeKey, err)
	}

	return m.convertMetricValueToString(value)
}

func (m *Metrics) GetMetricAsDTO(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error) {
	storeKey := m.buildStoreKey(name, metricType)
	value, err := m.store.Read(ctx, storeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read metric '%s': %w", storeKey, err)
	}

	return value, nil
}

func (m *Metrics) GetAllValuesAsString(ctx context.Context) (map[string]string, error) {
	dataToReturn := make(map[string]string)
	allMetrics, err := m.store.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read all metrics for output: %w", err)
	}

	for _, metricValue := range allMetrics {
		valueAsString, err := m.convertMetricValueToString(metricValue)
		if err != nil {
			return dataToReturn, err
		}
		dataToReturn[metricValue.ID] = valueAsString
	}

	return dataToReturn, nil
}

func (m *Metrics) convertMetricValueToString(metricDTO *dto.MetricDTO) (string, error) {
	switch metricDTO.MType {
	case common.CounterMetricType:
		if metricDTO.Delta == nil {
			return "", ErrCounterValueParse
		}
		return strconv.FormatInt(*metricDTO.Delta, 10), nil
	case common.GougeMetricType:
		spew.Dump(metricDTO)
		if metricDTO.Value == nil {
			return "", ErrGaugeValueParse
		}
		return strconv.FormatFloat(*metricDTO.Value, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
	}
}

func (m *Metrics) ExportMetrics(ctx context.Context) ([]*dto.MetricDTO, error) {
	allMetrics, err := m.store.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read all metrics for export: %w", err)
	}
	return allMetrics, nil
}

func (m *Metrics) ImportMetrics(ctx context.Context, metricsToImport []*dto.MetricDTO) error {
	for _, metricToImport := range metricsToImport {
		if err := m.SetMetricByDto(ctx, metricToImport); err != nil {
			return fmt.Errorf("failed to import metric '%s': %w", metricToImport.ID, err)
		}
	}
	return nil
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
	Create(ctx context.Context, value *dto.MetricDTO) error
	Read(ctx context.Context, key string) (*dto.MetricDTO, error)
	Update(ctx context.Context, value *dto.MetricDTO) error
	GetAll(ctx context.Context) ([]*dto.MetricDTO, error)
}

type Metrics struct {
	store Store
}

func NewMetrics(s Store) *Metrics {
	return &Metrics{store: s}
}
