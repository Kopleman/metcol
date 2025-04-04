package httpclient

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient_Post(t *testing.T) {
	testBody := []byte("test body")

	tests := []struct {
		name          string
		key           string
		mockResponse  *http.Response
		mockError     error
		expectedError string
		expectedHash  bool
		checkHeaders  bool
	}{
		{
			name: "successful request",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       gzipBody("test body"),
			},
			checkHeaders: true,
		},
		{
			name:         "request with HMAC hash",
			key:          "secret",
			mockResponse: &http.Response{StatusCode: http.StatusOK, Body: gzipBody("test body")},
			expectedHash: true,
			checkHeaders: true,
		},
		{
			name:          "server error response",
			mockResponse:  &http.Response{StatusCode: http.StatusInternalServerError},
			expectedError: "status code 500",
		},
		{
			name:          "network error",
			mockError:     errors.New("connection failed"),
			expectedError: "failed to send post req",
		},
		{
			name:          "invalid gzip response",
			mockResponse:  &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("invalid"))},
			expectedError: "failed to decompress response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRT := new(MockRoundTripper)
			mockLogger := new(log.MockLogger)

			// Configure client
			cfg := &config.Config{
				EndPoint: &flags.NetAddress{Host: "test-server", Port: "80"},
				Key:      tt.key,
			}
			client := NewHTTPClient(cfg, mockLogger)
			client.client.Transport = mockRT

			// Expected request validations
			mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).
				Return(tt.mockResponse, tt.mockError).
				Once()

			// Execute
			resp, err := client.Post("/test", "application/json", testBody)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
			require.NoError(t, err)

			if tt.checkHeaders {
				req := mockRT.Calls[0].Arguments[0].(*http.Request)
				assert.Equal(t, "application/json", req.Header.Get(common.ContentType))
				assert.Equal(t, "gzip", req.Header.Get(common.AcceptEncoding))
			}

			if tt.expectedHash {
				req := mockRT.Calls[0].Arguments[0].(*http.Request)
				expectedHash := calculateHash(testBody, []byte(tt.key))
				assert.Equal(t, expectedHash, req.Header.Get(common.HashSHA256))
			}

			if tt.mockResponse != nil && tt.mockResponse.StatusCode == http.StatusOK {
				assert.Equal(t, testBody, resp)
			}
		})
	}
}

func TestHTTPClient_CalcHashForBody(t *testing.T) {
	secret := []byte("secret-key")
	testBody := []byte("test-body")

	t.Run("with key and body", func(t *testing.T) {
		client := &HTTPClient{key: secret}
		hash := client.calcHashForBody(testBody)

		mac := hmac.New(sha256.New, secret)
		mac.Write(testBody)
		expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		assert.Equal(t, expected, hash)
	})

	t.Run("empty key", func(t *testing.T) {
		client := &HTTPClient{key: []byte{}}
		hash := client.calcHashForBody(testBody)
		assert.Empty(t, hash)
	})

	t.Run("empty body", func(t *testing.T) {
		client := &HTTPClient{key: secret}
		hash := client.calcHashForBody([]byte{})
		assert.Empty(t, hash)
	})
}

func TestNewHTTPClient(t *testing.T) {
	cfg := &config.Config{
		EndPoint: &flags.NetAddress{Host: "example.com", Port: "8080"},
		Key:      "test-key",
	}
	logger := new(log.MockLogger)

	client := NewHTTPClient(cfg, logger)

	assert.Equal(t, "http://example.com:8080", client.BaseURL)
	assert.NotNil(t, client.client.Transport)
	assert.Equal(t, []byte("test-key"), client.key)
}

func gzipBody(data string) io.ReadCloser {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(data))
	gz.Close()
	return io.NopCloser(&buf)
}

func calculateHash(body []byte, key []byte) string {
	if len(key) == 0 || len(body) == 0 {
		return ""
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
