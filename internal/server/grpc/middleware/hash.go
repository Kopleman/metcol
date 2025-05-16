package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const HashSHA256 = "HashSHA256"

func Hash(key []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if len(key) == 0 {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata not found")
		}

		hashValues := md.Get(HashSHA256)
		if len(hashValues) == 0 {
			return nil, status.Error(codes.InvalidArgument, "hash header not found")
		}

		receivedHash := hashValues[0]

		reqI, ok := req.(interface{ Marshal() ([]byte, error) })
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "request body not marshalling")
		}

		reqBytes, err := reqI.Marshal()
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal request")
		}

		h := hmac.New(sha256.New, key)
		h.Write(reqBytes)
		calculatedHash := hex.EncodeToString(h.Sum(nil))

		if receivedHash != calculatedHash {
			return nil, status.Error(codes.InvalidArgument, "invalid hash")
		}

		newCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(HashSHA256, calculatedHash))

		resp, err := handler(newCtx, req)
		if err != nil {
			return nil, err
		}

		respI, ok := resp.(interface{ Marshal() ([]byte, error) })
		if !ok {
			return nil, errors.New("response body not marshalling")
		}

		respBytes, err := respI.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		h = hmac.New(sha256.New, key)
		h.Write(respBytes)
		respHash := hex.EncodeToString(h.Sum(nil))

		if err = grpc.SetHeader(newCtx, metadata.Pairs(HashSHA256, respHash)); err != nil {
			return nil, status.Error(codes.Internal, "failed to set header")
		}

		return resp, nil
	}
}
