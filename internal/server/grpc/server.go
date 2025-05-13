package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/Kopleman/metcol/internal/common/log"
	grpcmiddleware "github.com/Kopleman/metcol/internal/server/grpc/middleware"
	pb "github.com/Kopleman/metcol/proto/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	server *grpc.Server
	logger log.Logger
}

func NewServer(logger log.Logger, metricsService *MetricsService, trustedCIDR string, key string) *Server {
	encoding.RegisterCompressor(encoding.GetCompressor("gzip"))

	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.IPFilter(trustedCIDR),
			grpcmiddleware.Hash([]byte(key)),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,
			MaxConnectionAge:      10 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  2 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	server := grpc.NewServer(serverOpts...)
	pb.RegisterMetricsServiceServer(server, metricsService)

	return &Server{
		server: server,
		logger: logger,
	}
}

func (s *Server) Start(grpcAddress string) error {
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Infof("gRPC server listening on %s", grpcAddress)
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
