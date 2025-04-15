package httpclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryableTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		expectedError    string
		mockResponses    []*http.Response
		mockErrors       []error
		retryCount       int
		expectedAttempts int
	}{
		{
			name:             "success on first attempt",
			retryCount:       3,
			mockResponses:    []*http.Response{{StatusCode: http.StatusOK}},
			mockErrors:       []error{nil},
			expectedAttempts: 1,
		},
		{
			name:             "retry on 5xx status",
			retryCount:       3,
			mockResponses:    []*http.Response{{StatusCode: http.StatusInternalServerError}, {StatusCode: http.StatusOK}},
			mockErrors:       []error{errors.New("some error"), nil},
			expectedAttempts: 2,
		},
		{
			name:             "retry on connection error",
			retryCount:       3,
			mockResponses:    []*http.Response{nil, nil, {StatusCode: http.StatusOK}},
			mockErrors:       []error{errors.New("connection error"), errors.New("connection error"), nil},
			expectedAttempts: 3,
		},
		{
			name:             "exceed max retries",
			retryCount:       2,
			mockResponses:    []*http.Response{nil, nil, nil},
			mockErrors:       []error{errors.New("error 1"), errors.New("error 2"), errors.New("error 3")},
			expectedAttempts: 3,
			expectedError:    "retry amount exeeded: error 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := new(MockRoundTripper)
			mockLogger := new(log.MockLogger)

			transport := &retryableTransport{
				transport:  mockRT,
				logger:     mockLogger,
				retryCount: tt.retryCount,
			}

			req, _ := http.NewRequest(http.MethodGet, "http://test.com", bytes.NewBufferString("test body"))

			// Setup mock calls
			for i := range len(tt.mockResponses) {
				mockRT.On("RoundTrip", req).Return(tt.mockResponses[i], tt.mockErrors[i]).Once()
			}

			resp, err := transport.RoundTrip(req)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close() //nolint:all // tests
			}

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponses[tt.expectedAttempts-1], resp)
			}

			mockRT.AssertNumberOfCalls(t, "RoundTrip", tt.expectedAttempts)
		})
	}
}

func TestBackoff(t *testing.T) {
	tests := []struct {
		retries int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("retries %d", tt.retries), func(t *testing.T) {
			got := backoff(tt.retries)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		err      error
		resp     *http.Response
		name     string
		expected bool
	}{
		{
			name:     "error occurred",
			err:      errors.New("connection error"),
			expected: true,
		},
		{
			name:     "5xx status",
			resp:     &http.Response{StatusCode: http.StatusInternalServerError},
			expected: true,
		},
		{
			name:     "4xx status",
			resp:     &http.Response{StatusCode: http.StatusBadRequest},
			expected: false,
		},
		{
			name:     "2xx status",
			resp:     &http.Response{StatusCode: http.StatusOK},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetry(tt.err, tt.resp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloseBody(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString("test"))}
		err := closeBody(resp)
		assert.NoError(t, err)
	})

	t.Run("nil body", func(t *testing.T) {
		resp := &http.Response{}
		err := closeBody(resp)
		assert.NoError(t, err)
	})

	t.Run("close error", func(t *testing.T) {
		resp := &http.Response{Body: &errorReaderCloser{}}
		err := closeBody(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to close response body")
	})
}

type errorReaderCloser struct{}

func (e *errorReaderCloser) Read(p []byte) (n int, err error) { return 0, io.EOF }
func (e *errorReaderCloser) Close() error                     { return errors.New("close error") }

func TestBodyReuse(t *testing.T) {
	mockRT := new(MockRoundTripper)
	mockLogger := new(log.MockLogger)

	transport := &retryableTransport{
		transport:  mockRT,
		logger:     mockLogger,
		retryCount: 2,
	}

	originalBody := []byte("original body")
	req, _ := http.NewRequest(http.MethodPost, "http://test.com", bytes.NewBuffer(originalBody))

	// First attempt fails
	mockRT.On("RoundTrip", req).Return(
		&http.Response{StatusCode: http.StatusInternalServerError},
		errors.New("connection error"),
	).Once()
	// Second attempt succeeds
	mockRT.On("RoundTrip", req).Return(&http.Response{StatusCode: http.StatusOK}, nil).Once()

	resp, err := transport.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //nolint:all // tests
	}
	require.NoError(t, err)

	// Verify body can be read again
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, originalBody, body)

	mockRT.AssertExpectations(t)
}

func TestNewRetryableTransport(t *testing.T) {
	logger := new(log.MockLogger)
	retryCount := 3

	transport := NewRetryableTransport(logger, retryCount).(*retryableTransport) //nolint:all //tests

	assert.NotNil(t, transport.transport)
	assert.Equal(t, retryCount, transport.retryCount)
	assert.Equal(t, logger, transport.logger)
}
