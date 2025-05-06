package grpc

import (
	"fmt"
	"net"

	"github.com/Kopleman/metcol/internal/common/log"
	pb "github.com/Kopleman/metcol/proto/metrics"
	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	logger log.Logger
}

func NewServer(logger log.Logger, metricsService *MetricsService) *Server {
	server := grpc.NewServer()
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
