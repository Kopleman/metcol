package utils

import (
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	pb "github.com/Kopleman/metcol/proto/metrics"
)

func ConvertProtoMetricType(t pb.MetricType) common.MetricType {
	switch t {
	case pb.MetricType_GAUGE:
		return common.GaugeMetricType
	case pb.MetricType_COUNTER:
		return common.CounterMetricType
	default:
		return common.UnknownMetricType
	}
}

func ConvertProtoMetricToDTO(m *pb.Metric) *dto.MetricDTO {
	metric := &dto.MetricDTO{
		ID:    m.Id,
		MType: ConvertProtoMetricType(m.Type),
	}

	if m.Type == pb.MetricType_GAUGE {
		metric.Value = &m.Value
	} else if m.Type == pb.MetricType_COUNTER {
		metric.Delta = &m.Delta
	}

	return metric
}

func ConvertDTOToProtoMetric(m *dto.MetricDTO) *pb.Metric {
	metric := &pb.Metric{
		Id:   m.ID,
		Type: ConvertDTOMetricType(m.MType.String()),
	}

	if m.Value != nil {
		metric.Value = *m.Value
	}
	if m.Delta != nil {
		metric.Delta = *m.Delta
	}

	return metric
}

func ConvertDTOMetricType(t string) pb.MetricType {
	switch t {
	case common.GaugeMetricType.String():
		return pb.MetricType_GAUGE
	case common.CounterMetricType.String():
		return pb.MetricType_COUNTER
	default:
		return pb.MetricType_UNKNOWN
	}
}
