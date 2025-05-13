package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	pb "github.com/Kopleman/metcol/proto/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsClient struct {
	client pb.MetricsServiceClient
	conn   *grpc.ClientConn
	key    []byte
}

func NewMetricsClient(address string, key string) (*MetricsClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy": "round_robin",
			"methodConfig": [{
				"name": [{"service": "metrics.MetricsService"}],
				"waitForReady": true,
				"compression": "gzip",
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "3s",
					"backoffMultiplier": 1.6,
					"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
				}
			}]
		}`),
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
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	client := pb.NewMetricsServiceClient(conn)
	return &MetricsClient{
		client: client,
		conn:   conn,
		key:    []byte(key),
	}, nil
}

func (c *MetricsClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

func (c *MetricsClient) addHashToContext(ctx context.Context, req interface{}) (context.Context, error) {
	if len(c.key) == 0 {
		return ctx, nil
	}

	reqI, ok := req.(interface{ Marshal() ([]byte, error) })
	if !ok {
		return nil, fmt.Errorf(
			"request body not marshalling: %w",
			status.Error(codes.InvalidArgument, "request body not marshalling"),
		)
	}
	reqBytes, err := reqI.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	h := hmac.New(sha256.New, c.key)
	h.Write(reqBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	return metadata.AppendToOutgoingContext(ctx, "HashSHA256", hash), nil
}

func (c *MetricsClient) GetMetric(ctx context.Context, id string, metricType pb.MetricType) (*pb.Metric, error) {
	req := &pb.GetMetricRequest{}
	req.SetId(id)
	req.SetType(metricType)

	ctx, err := c.addHashToContext(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.GetMetric(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric %s: %w", id, err)
	}

	return resp.GetMetric(), nil
}

func (c *MetricsClient) UpdateMetric(ctx context.Context, metric *pb.Metric) (*pb.Metric, error) {
	req := &pb.UpdateMetricRequest{}
	req.SetMetric(metric)

	ctx, err := c.addHashToContext(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.UpdateMetric(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update metric %s: %w", metric, err)
	}

	return resp.GetMetric(), nil
}

func (c *MetricsClient) UpdateMetrics(ctx context.Context, metrics []*pb.Metric) ([]*pb.Metric, error) {
	req := &pb.UpdateMetricsRequest{}
	req.SetMetrics(metrics)

	ctx, err := c.addHashToContext(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update metrics: %w", err)
	}

	return resp.GetMetrics(), nil
}

func (c *MetricsClient) GetAllMetrics(ctx context.Context) ([]*pb.Metric, error) {
	req := &pb.GetAllMetricsRequest{}

	ctx, err := c.addHashToContext(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.GetAllMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}

	return resp.GetMetrics(), nil
}
