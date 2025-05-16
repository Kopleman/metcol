package middleware

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func IPFilter(trustedCIDR string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "peer information not available")
		}

		clientIP, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid client address")
		}

		_, ipnet, err := net.ParseCIDR(trustedCIDR)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid CIDR format")
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid IP address format")
		}

		if !ipnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		return handler(ctx, req)
	}
}
