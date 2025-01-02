package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const baseBackoffMultiplier = 2

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(baseBackoffMultiplier, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return true
	}
	return false
}

func drainBody(resp *http.Response) error {
	if resp != nil && resp.Body != nil {
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return fmt.Errorf("failed to drain response body: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("failed to close response body: %w", err)
		}
	}

	return nil
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	resp, err := t.transport.RoundTrip(req)
	retries := 0
	for shouldRetry(err, resp) && retries < t.retryCount {
		time.Sleep(backoff(retries))
		if err = drainBody(resp); err != nil {
			return nil, fmt.Errorf("round trip: %w", err)
		}
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		resp, err = t.transport.RoundTrip(req)
		retries++
	}
	return resp, fmt.Errorf("retry amount exeeded: %w", err)
}

type retryableTransport struct {
	transport  http.RoundTripper
	retryCount int
}

func NewRetryableTransport(retryCount int) http.RoundTripper {
	transport := &retryableTransport{
		transport:  &http.Transport{},
		retryCount: retryCount,
	}

	return transport
}
