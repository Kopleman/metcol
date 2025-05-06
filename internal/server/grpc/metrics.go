package grpc

import (
	"context"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
	pb "github.com/Kopleman/metcol/proto/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *MetricsService) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	metricType := utils.ConvertProtoMetricType(req.Type)
	metric, err := s.metricsService.GetMetricAsDTO(ctx, metricType, req.Id)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to get metric")
	}

	return &pb.GetMetricResponse{
		Metric: utils.ConvertDTOToProtoMetric(metric),
	}, nil
}

func (s *MetricsService) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	metricDto := utils.ConvertProtoMetricToDTO(req.Metric)
	err := s.metricsService.SetMetricByDto(ctx, metricDto)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to update metric")
	}

	return &pb.UpdateMetricResponse{
		Metric: req.Metric,
	}, nil
}

func (s *MetricsService) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]*dto.MetricDTO, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		metrics = append(metrics, utils.ConvertProtoMetricToDTO(m))
	}

	err := s.metricsService.SetMetrics(ctx, metrics)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &pb.UpdateMetricsResponse{
		Metrics: req.Metrics,
	}, nil
}

func (s *MetricsService) GetAllMetrics(ctx context.Context, req *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	allMetrics, err := s.metricsService.ExportMetrics(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to get all metrics")
	}

	metrics := make([]*pb.Metric, 0, len(allMetrics))
	for _, m := range allMetrics {
		metrics = append(metrics, utils.ConvertDTOToProtoMetric(m))
	}

	return &pb.GetAllMetricsResponse{
		Metrics: metrics,
	}, nil
}
