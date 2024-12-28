package pgxstore

import (
	"errors"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
)

func (m *Metric) convertMetricType() (common.MetricType, error) {
	if m == nil {
		return common.UnknownMetricType, errors.New("nil metric")
	}

	switch m.Type {
	case MetricTypeGauge:
		return common.GaugeMetricType, nil
	case MetricTypeCounter:
		return common.CounterMetricType, nil
	default:
		return common.UnknownMetricType, fmt.Errorf("unknown metric type: %v", m.Type)
	}
}

func (m *Metric) toDTO() (*dto.MetricDTO, error) {
	if m == nil {
		return nil, errors.New("nil metric")
	}
	mType, err := m.convertMetricType()
	if err != nil {
		return nil, fmt.Errorf("dto converting error: %w", err)
	}

	metricDto := dto.MetricDTO{
		ID:    m.Name,
		MType: mType,
		Delta: m.Delta,
		Value: m.Value,
	}

	return &metricDto, nil
}
