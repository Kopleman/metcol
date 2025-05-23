package utils

import (
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	pb "github.com/Kopleman/metcol/proto/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConvertProtoMetricType(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.MetricType
		expected common.MetricType
	}{
		{
			name:     "Convert GAUGE",
			input:    pb.MetricType_GAUGE,
			expected: common.GaugeMetricType,
		},
		{
			name:     "Convert COUNTER",
			input:    pb.MetricType_COUNTER,
			expected: common.CounterMetricType,
		},
		{
			name:     "Convert UNKNOWN",
			input:    pb.MetricType_UNKNOWN,
			expected: common.UnknownMetricType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertProtoMetricType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertProtoMetricToDTO(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.Metric
		expected *dto.MetricDTO
	}{
		{
			name: "Convert GAUGE metric",
			input: func() *pb.Metric {
				m := &pb.Metric{}
				m.SetId("test_gauge")
				m.SetType(pb.MetricType_GAUGE)
				m.SetValue(42.5)
				return m
			}(),
			expected: &dto.MetricDTO{
				ID:    "test_gauge",
				MType: common.GaugeMetricType,
				Value: func() *float64 { v := 42.5; return &v }(),
			},
		},
		{
			name: "Convert COUNTER metric",
			input: func() *pb.Metric {
				m := &pb.Metric{}
				m.SetId("test_counter")
				m.SetType(pb.MetricType_COUNTER)
				m.SetDelta(100)
				return m
			}(),
			expected: &dto.MetricDTO{
				ID:    "test_counter",
				MType: common.CounterMetricType,
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertProtoMetricToDTO(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertDTOToProtoMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    *dto.MetricDTO
		expected *pb.Metric
	}{
		{
			name: "Convert GAUGE metric",
			input: &dto.MetricDTO{
				ID:    "test_gauge",
				MType: common.GaugeMetricType,
				Value: func() *float64 { v := 42.5; return &v }(),
			},
			expected: func() *pb.Metric {
				m := &pb.Metric{}
				m.SetId("test_gauge")
				m.SetType(pb.MetricType_GAUGE)
				m.SetValue(42.5)
				return m
			}(),
		},
		{
			name: "Convert COUNTER metric",
			input: &dto.MetricDTO{
				ID:    "test_counter",
				MType: common.CounterMetricType,
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
			expected: func() *pb.Metric {
				m := &pb.Metric{}
				m.SetId("test_counter")
				m.SetType(pb.MetricType_COUNTER)
				m.SetDelta(100)
				return m
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertDTOToProtoMetric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertDTOMetricType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected pb.MetricType
	}{
		{
			name:     "Convert GAUGE string",
			input:    common.GaugeMetricType.String(),
			expected: pb.MetricType_GAUGE,
		},
		{
			name:     "Convert COUNTER string",
			input:    common.CounterMetricType.String(),
			expected: pb.MetricType_COUNTER,
		},
		{
			name:     "Convert unknown string",
			input:    "unknown",
			expected: pb.MetricType_UNKNOWN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertDTOMetricType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
