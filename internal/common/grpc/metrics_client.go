package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Kopleman/metcol/proto/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type MetricsClient struct {
	client pb.MetricsServiceClient
	conn   *grpc.ClientConn
}

// NewMetricsClient создает нового клиента для работы с метриками
func NewMetricsClient(address string) (*MetricsClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   3 * time.Second,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
	}

	conn, err := grpc.NewClient(
		address,
		dialOpts...,
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewMetricsServiceClient(conn)
	return &MetricsClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close закрывает соединение с сервером
func (c *MetricsClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

// GetMetric получает значение метрики
func (c *MetricsClient) GetMetric(ctx context.Context, id string, metricType pb.MetricType) (*pb.Metric, error) {
	req := &pb.GetMetricRequest{
		Id:   id,
		Type: metricType,
	}

	resp, err := c.client.GetMetric(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Metric, nil
}

// UpdateMetric обновляет значение метрики
func (c *MetricsClient) UpdateMetric(ctx context.Context, metric *pb.Metric) (*pb.Metric, error) {
	req := &pb.UpdateMetricRequest{
		Metric: metric,
	}

	resp, err := c.client.UpdateMetric(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Metric, nil
}

// UpdateMetrics обновляет несколько метрик
func (c *MetricsClient) UpdateMetrics(ctx context.Context, metrics []*pb.Metric) ([]*pb.Metric, error) {
	req := &pb.UpdateMetricsRequest{
		Metrics: metrics,
	}

	resp, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Metrics, nil
}

// GetAllMetrics получает все метрики
func (c *MetricsClient) GetAllMetrics(ctx context.Context) ([]*pb.Metric, error) {
	req := &pb.GetAllMetricsRequest{}

	resp, err := c.client.GetAllMetrics(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Metrics, nil
}
