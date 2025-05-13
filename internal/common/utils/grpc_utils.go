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
		ID:    m.GetId(),
		MType: ConvertProtoMetricType(m.GetType()),
	}

	if m.GetType() == pb.MetricType_GAUGE {
		value := m.GetValue()
		metric.Value = &value
	} else if m.GetType() == pb.MetricType_COUNTER {
		delta := m.GetDelta()
		metric.Delta = &delta
	}

	return metric
}

func ConvertDTOToProtoMetric(m *dto.MetricDTO) *pb.Metric {
	metric := &pb.Metric{}
	metric.SetId(m.ID)
	metric.SetType(ConvertDTOMetricType(m.MType.String()))

	if m.Value != nil {
		metric.SetValue(*m.Value)
	}
	if m.Delta != nil {
		metric.SetDelta(*m.Delta)
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
