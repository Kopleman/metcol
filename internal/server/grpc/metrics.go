package grpc

import (
	"context"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
	pb "github.com/Kopleman/metcol/proto/metrics"
)

type MetricsService struct {
	pb.UnimplementedMetricsServiceServer
	logger         log.Logger
	metricsService Metrics
}

type Metrics interface {
	SetMetric(ctx context.Context, metricType common.MetricType, name string, value string) error
	SetMetricByDto(ctx context.Context, metricDto *dto.MetricDTO) error
	GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error)
	GetAllValuesAsString(ctx context.Context) (map[string]string, error)
	SetMetrics(ctx context.Context, metrics []*dto.MetricDTO) error
	ExportMetrics(ctx context.Context) ([]*dto.MetricDTO, error)
}

func NewMetricsService(logger log.Logger, metricsService Metrics) *MetricsService {
	return &MetricsService{
		logger:         logger,
		metricsService: metricsService,
	}
}

func (s *MetricsService) GetMetric(
	ctx context.Context,
	req *pb.GetMetricRequest,
) (*pb.GetMetricResponse, error) {
	metricType := utils.ConvertProtoMetricType(req.GetType())
	metric, err := s.metricsService.GetMetricAsDTO(ctx, metricType, req.GetId())
	if err != nil {
		s.logger.Error(err)
		return nil, fmt.Errorf("unable to get metric: %w", err)
	}

	resp := &pb.GetMetricResponse{}
	resp.SetMetric(utils.ConvertDTOToProtoMetric(metric))
	return resp, nil
}

func (s *MetricsService) UpdateMetric(
	ctx context.Context,
	req *pb.UpdateMetricRequest,
) (*pb.UpdateMetricResponse, error) {
	metricDto := utils.ConvertProtoMetricToDTO(req.GetMetric())
	err := s.metricsService.SetMetricByDto(ctx, metricDto)
	if err != nil {
		s.logger.Error(err)
		return nil, fmt.Errorf("unable to update metric: %w", err)
	}

	resp := &pb.UpdateMetricResponse{}
	resp.SetMetric(req.GetMetric())
	return resp, nil
}

func (s *MetricsService) UpdateMetrics(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]*dto.MetricDTO, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		metrics = append(metrics, utils.ConvertProtoMetricToDTO(m))
	}

	err := s.metricsService.SetMetrics(ctx, metrics)
	if err != nil {
		s.logger.Error(err)
		return nil, fmt.Errorf("unable to update metrics: %w", err)
	}

	resp := &pb.UpdateMetricsResponse{}
	resp.SetMetrics(req.GetMetrics())
	return resp, nil
}

func (s *MetricsService) GetAllMetrics(
	ctx context.Context,
	req *pb.GetAllMetricsRequest,
) (*pb.GetAllMetricsResponse, error) {
	allMetrics, err := s.metricsService.ExportMetrics(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, fmt.Errorf("unable to get metrics: %w", err)
	}

	metrics := make([]*pb.Metric, 0, len(allMetrics))
	for _, m := range allMetrics {
		metrics = append(metrics, utils.ConvertDTOToProtoMetric(m))
	}

	resp := &pb.GetAllMetricsResponse{}
	resp.SetMetrics(metrics)
	return resp, nil
}
