package metrics

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/sterrors"
	"github.com/Kopleman/metcol/internal/server/store"
)

func (m *Metrics) SetGauge(ctx context.Context, name string, value float64) (*float64, error) {
	_, err := m.store.Read(ctx, common.GaugeMetricType, name)

	metricDTO := &dto.MetricDTO{
		Delta: nil,
		Value: &value,
		ID:    name,
		MType: common.GaugeMetricType,
	}
	if err != nil {
		if errors.Is(err, sterrors.ErrNotFound) {
			storeErr := m.store.Create(ctx, metricDTO)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create gauge metric '%s': %w", name, storeErr)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read gauge metric '%s': %w", name, err)
	}

	updateErr := m.store.Update(ctx, metricDTO)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update gauge metric '%s': %w", name, err)
	}

	return &value, nil
}

func (m *Metrics) SetCounter(ctx context.Context, name string, value int64) (*int64, error) {
	existedCounter, err := m.store.Read(ctx, common.CounterMetricType, name)

	if err != nil {
		if errors.Is(err, sterrors.ErrNotFound) {
			metricDTO := &dto.MetricDTO{
				Delta: &value,
				Value: nil,
				ID:    name,
				MType: common.CounterMetricType,
			}
			storeErr := m.store.Create(ctx, metricDTO)
			if storeErr != nil {
				return nil, fmt.Errorf("failed to create counter metric '%s': %w", name, err)
			}
			return &value, nil
		}

		return nil, fmt.Errorf("failed to read counter metric '%s': %w", name, err)
	}

	if existedCounter.Delta == nil {
		return nil, ErrCounterValueParse
	}

	newValue := *existedCounter.Delta + value
	existedCounter.Delta = &newValue
	updateErr := m.store.Update(ctx, existedCounter)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update counter metric '%s': %w", name, err)
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
	case common.GaugeMetricType:
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

func (m *Metrics) SetMetrics(ctx context.Context, metricDTOs []*dto.MetricDTO) error {
	dtoForSet, err := m.prepareMetricDTOForSet(ctx, metricDTOs)
	if err != nil {
		return fmt.Errorf("metrics.SetMetrics prepare metric DTOs for set: %w", err)
	}

	if err = m.store.BulkCreateOrUpdate(ctx, dtoForSet); err != nil {
		return fmt.Errorf("metrics.setMetric BulkCreateOrUpdate: %w", err)
	}

	return nil
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
	case common.GaugeMetricType:
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

func (m *Metrics) validateMetricDto(d *dto.MetricDTO) error {
	switch d.MType {
	case common.CounterMetricType:
		if d.Delta == nil {
			return ErrValueParse
		}
		return nil
	case common.GaugeMetricType:
		if d.Value == nil {
			return ErrValueParse
		}
		return nil
	default:
		return ErrUnknownMetricType
	}
}

func (m *Metrics) prepareMetricDTOForSet(ctx context.Context, metricDTOs []*dto.MetricDTO) ([]*dto.MetricDTO, error) {
	dtoForSet := make([]*dto.MetricDTO, 0)
	for _, d := range metricDTOs {
		if err := m.validateMetricDto(d); err != nil {
			return nil, fmt.Errorf("metrics.prepareDataForSet validate input: %w", err)
		}

		// check that we did not meet this metric before.
		indexOfPrepared := slices.IndexFunc(dtoForSet,
			func(preparedDTO *dto.MetricDTO) bool {
				return preparedDTO.MType == d.MType && preparedDTO.ID == d.ID
			})

		var existedMetric *dto.MetricDTO
		if indexOfPrepared != -1 {
			existedMetric = dtoForSet[indexOfPrepared]
		}

		if existedMetric == nil {
			metricInStore, readErr := m.store.Read(ctx, common.CounterMetricType, d.ID)
			if readErr != nil {
				if errors.Is(readErr, sterrors.ErrNotFound) {
					dtoForSet = append(dtoForSet, d)
					continue
				}
				return nil, fmt.Errorf("metrics.prepareDataForSet read: %w", readErr)
			}
			existedMetric = metricInStore
			dtoForSet = append(dtoForSet, existedMetric)
		}

		if err := m.validateMetricDto(existedMetric); err != nil {
			return nil, fmt.Errorf("metrics.prepareDataForSet validate existed: %w", err)
		}

		switch existedMetric.MType {
		case common.CounterMetricType:
			newValue := *existedMetric.Delta + *d.Delta
			existedMetric.Delta = &newValue
		case common.GaugeMetricType:
			existedMetric.Value = d.Value
		default:
			return nil, ErrUnknownMetricType
		}
	}

	return dtoForSet, nil
}

func (m *Metrics) GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error) {
	value, err := m.store.Read(ctx, metricType, name)
	if err != nil {
		return "", fmt.Errorf("failed to read metric '%s': %w", name, err)
	}

	return m.convertMetricValueToString(value)
}

func (m *Metrics) GetMetricAsDTO(
	ctx context.Context,
	metricType common.MetricType,
	name string) (*dto.MetricDTO, error) {
	value, err := m.store.Read(ctx, metricType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read metric '%s': %w", name, err)
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
	case common.GaugeMetricType:
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
	case string(common.GaugeMetricType):
		return common.GaugeMetricType, nil
	default:
		return common.UnknownMetricType, ErrUnknownMetricType
	}
}

type Metrics struct {
	store  store.Store
	logger log.Logger
}

func NewMetrics(s store.Store, logger log.Logger) *Metrics {
	return &Metrics{store: s, logger: logger}
}
