package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestIPFilter(t *testing.T) {
	tests := []struct {
		name        string
		trustedCIDR string
		clientIP    string
		wantErr     bool
		errCode     codes.Code
	}{
		{
			name:        "valid IP in trusted subnet",
			trustedCIDR: "192.168.1.0/24",
			clientIP:    "192.168.1.100",
			wantErr:     false,
		},
		{
			name:        "IP not in trusted subnet",
			trustedCIDR: "192.168.1.0/24",
			clientIP:    "10.0.0.1",
			wantErr:     true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "invalid CIDR format",
			trustedCIDR: "invalid-cidr",
			clientIP:    "192.168.1.100",
			wantErr:     true,
			errCode:     codes.Internal,
		},
		{
			name:        "invalid IP format",
			trustedCIDR: "192.168.1.0/24",
			clientIP:    "invalid-ip",
			wantErr:     true,
			errCode:     codes.InvalidArgument,
		},
		{
			name:        "empty trusted CIDR",
			trustedCIDR: "",
			clientIP:    "192.168.1.100",
			wantErr:     true,
			errCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := IPFilter(tt.trustedCIDR)

			// Создаем тестовый контекст с peer информацией
			ctx := peer.NewContext(context.Background(), &peer.Peer{
				Addr: &mockAddr{addr: tt.clientIP + ":12345"},
			})

			// Создаем тестовый обработчик
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			// Вызываем middleware
			resp, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{}, handler)

			if tt.wantErr {
				require.Error(t, err)
				if st, ok := status.FromError(err); ok {
					require.Equal(t, tt.errCode, st.Code())
				}
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, "success", resp)
			}
		})
	}
}

// mockAddr реализует net.Addr для тестов
type mockAddr struct {
	addr string
}

func (m *mockAddr) Network() string {
	return "tcp"
}

func (m *mockAddr) String() string {
	return m.addr
}
