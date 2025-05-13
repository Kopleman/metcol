package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockProtoMessage реализует proto.Message для тестов
type mockProtoMessage struct {
	data []byte
}

func (m *mockProtoMessage) Marshal() ([]byte, error) {
	if m.data == nil {
		return nil, errors.New("marshal error")
	}
	return m.data, nil
}

func calculateHash(key []byte, data []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func TestHash(t *testing.T) {
	testData := []byte("test-data")
	testKey := []byte("test-key")
	validHash := calculateHash(testKey, testData)

	tests := []struct {
		req       *mockProtoMessage
		name      string
		reqHash   string
		key       []byte
		errCode   codes.Code
		wantErr   bool
		checkResp bool
	}{
		{
			name:      "valid hash",
			key:       testKey,
			req:       &mockProtoMessage{data: testData},
			reqHash:   validHash,
			wantErr:   false,
			checkResp: true,
		},
		{
			name:    "invalid hash",
			key:     testKey,
			req:     &mockProtoMessage{data: testData},
			reqHash: "invalid-hash",
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name:    "missing hash",
			key:     testKey,
			req:     &mockProtoMessage{data: testData},
			reqHash: "",
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name:    "empty key",
			key:     []byte{},
			req:     &mockProtoMessage{data: testData},
			reqHash: "any-hash",
			wantErr: false,
		},
		{
			name:    "marshaling error",
			key:     testKey,
			req:     &mockProtoMessage{data: nil},
			reqHash: "any-hash",
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := Hash(tt.key)

			// Создаем контекст с метаданными
			ctx := context.Background()
			if tt.reqHash != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(common.HashSHA256, tt.reqHash))
			}

			// Создаем тестовый обработчик
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return &mockProtoMessage{data: []byte("response-data")}, nil
			}

			// Вызываем middleware
			resp, err := interceptor(ctx, tt.req, &grpc.UnaryServerInfo{}, handler)

			if tt.wantErr {
				require.Error(t, err)
				if st, ok := status.FromError(err); ok {
					require.Equal(t, tt.errCode, st.Code())
				}
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Проверяем, что ответ содержит хэш в метаданных
				if tt.checkResp {
					md, ok := metadata.FromIncomingContext(ctx)
					require.True(t, ok)
					hashValues := md.Get(common.HashSHA256)
					require.Len(t, hashValues, 1)
					require.NotEmpty(t, hashValues[0])

					// Проверяем корректность хэша запроса
					reqData, err := tt.req.Marshal()
					require.NoError(t, err)
					expectedReqHash := calculateHash(tt.key, reqData)
					require.Equal(t, expectedReqHash, hashValues[0])
				}
			}
		})
	}
}

func TestHash_ResponseHash(t *testing.T) {
	testKey := []byte("test-key")
	testData := []byte("test-data")
	respData := []byte("response-data")
	validHash := calculateHash(testKey, testData)

	interceptor := Hash(testKey)

	// Создаем контекст с метаданными
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(common.HashSHA256, validHash))

	// Создаем тестовый обработчик
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &mockProtoMessage{data: respData}, nil
	}

	// Вызываем middleware
	resp, err := interceptor(ctx, &mockProtoMessage{data: testData}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Проверяем, что в контексте есть хэш ответа
	md, ok := metadata.FromIncomingContext(ctx)
	require.True(t, ok)
	hashValues := md.Get(common.HashSHA256)
	require.Len(t, hashValues, 1)
	require.NotEmpty(t, hashValues[0])

	// Проверяем корректность хэша запроса
	expectedReqHash := calculateHash(testKey, testData)
	require.Equal(t, expectedReqHash, hashValues[0])
}
